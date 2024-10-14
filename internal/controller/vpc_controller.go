/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

// import (
// 	"context"

// 	"k8s.io/apimachinery/pkg/runtime"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/log"

// 	vpcv1 "github.com/SMJNayeem/aws-vpc-controller/api/v1"
// )

// // VPCReconciler reconciles a VPC object
// type VPCReconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// }

// // +kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs,verbs=get;list;watch;create;update;patch;delete
// // +kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs/status,verbs=get;update;patch
// // +kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs/finalizers,verbs=update

// // Reconcile is part of the main kubernetes reconciliation loop which aims to
// // move the current state of the cluster closer to the desired state.
// // TODO(user): Modify the Reconcile function to compare the state specified by
// // the VPC object against the actual cluster state, and then
// // perform operations to make the cluster state reflect the state specified by
// // the user.
// //
// // For more details, check Reconcile and its Result here:
// // - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vpcv1 "github.com/SMJNayeem/aws-vpc-controller/api/v1"
)

// VPCReconciler reconciles a VPC object
type VPCReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpc.poridhi.io,resources=vpcs/finalizers,verbs=update

func (r *VPCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the VPC instance
	var vpc vpcv1.VPC
	if err := r.Get(ctx, req.NamespacedName, &vpc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(vpc.Spec.Region),
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create EC2 service client
	svc := ec2.New(sess)

	// Check if VPC exists
	if vpc.Status.VPCID == "" {
		// Create new VPC
		createVpcInput := &ec2.CreateVpcInput{
			CidrBlock: aws.String(vpc.Spec.CIDR),
		}
		createVpcOutput, err := svc.CreateVpc(createVpcInput)
		if err != nil {
			log.Error(err, "Failed to create VPC")
			vpc.Status.Error = err.Error()
			r.Status().Update(ctx, &vpc)
			return ctrl.Result{}, err
		}

		vpc.Status.VPCID = *createVpcOutput.Vpc.VpcId
		vpc.Status.State = *createVpcOutput.Vpc.State
		r.Status().Update(ctx, &vpc)

		log.Info("Created VPC", "VPCID", vpc.Status.VPCID)
	} else {
		// Update existing VPC
		describeVpcInput := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(vpc.Status.VPCID)},
		}
		describeVpcOutput, err := svc.DescribeVpcs(describeVpcInput)
		if err != nil {
			log.Error(err, "Failed to describe VPC")
			return ctrl.Result{}, err
		}

		if len(describeVpcOutput.Vpcs) == 0 {
			log.Info("VPC not found, recreating")
			vpc.Status.VPCID = ""
			r.Status().Update(ctx, &vpc)
			return ctrl.Result{Requeue: true}, nil
		}

		existingVpc := describeVpcOutput.Vpcs[0]
		if *existingVpc.CidrBlock != vpc.Spec.CIDR {
			// CIDR block has changed, delete and recreate VPC
			deleteVpcInput := &ec2.DeleteVpcInput{
				VpcId: aws.String(vpc.Status.VPCID),
			}
			_, err := svc.DeleteVpc(deleteVpcInput)
			if err != nil {
				log.Error(err, "Failed to delete VPC")
				return ctrl.Result{}, err
			}

			vpc.Status.VPCID = ""
			r.Status().Update(ctx, &vpc)
			return ctrl.Result{Requeue: true}, nil
		}

		vpc.Status.State = *existingVpc.State
		r.Status().Update(ctx, &vpc)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VPCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vpcv1.VPC{}).
		Complete(r)
}
