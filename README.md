# AWS VPC Controller

This project implements a simple Kubernetes controller for managing AWS Virtual Private Clouds (VPCs) using the Operator pattern. It is built using the Kubebuilder framework and allows you to create, update, and delete AWS VPCs directly through Kubernetes Custom Resources.

## Table of Contents

1. [Basic Terms](#basic-terms)
2. [Architecture Overview](#architecture-overview)
3. [System Flow and Dependencies](#system-flow-and-dependencies)
4. [Prerequisites](#prerequisites)
5. [Setup and Installation](#setup-and-installation)
6. [Usage](#usage)
7. [Development](#development)

## Basic Terms

- **Kubernetes**: An open-source container orchestration platform for automating deployment, scaling, and management of containerized applications.
- **Controller**: In Kubernetes, a controller is a control loop that watches the state of your cluster, then makes or requests changes where needed.
- **Custom Resource Definition (CRD)**: Allows you to define custom resources in Kubernetes.
- **Kubebuilder**: A framework for building Kubernetes APIs using custom resource definitions (CRDs).
- **AWS VPC**: Amazon Virtual Private Cloud, a virtual network dedicated to your AWS account.
- **Kind**: A tool for running local Kubernetes clusters using Docker containers as "nodes".

## Architecture Overview

![Architecture Overview](<images/Screenshot from 2024-10-15 00-30-02.png>)

This diagram illustrates the high-level architecture of the AWS VPC Controller:

1. The AWS VPC Controller runs within a Kubernetes cluster.
2. It watches for changes to VPC Custom Resources.
3. When changes are detected, it interacts with AWS using the AWS SDK to create, update, or delete VPCs.
4. Users interact with the system by creating, updating, or deleting VPC Custom Resources using kubectl.

## System Flow and Dependencies

![System Flow and Dependencies](<images/Screenshot from 2024-10-15 00-30-21.png>)

This sequence diagram shows the flow of operations and dependencies in the system:

1. The user applies a VPC Custom Resource (CR) using kubectl.
2. The Kubernetes API Server receives the CR and notifies the VPC Controller.
3. The VPC Controller processes the CR and uses the AWS SDK to interact with AWS.
4. AWS creates or updates the VPC as requested.
5. The VPC Controller updates the status of the CR with the results from AWS.
6. The user can see the updated status of their VPC CR.

## Prerequisites

- Go 1.20+
- Docker
- kubectl
- Kind
- AWS CLI configured with appropriate credentials
- Kubebuilder

## Setup and Installation

1. Clone the repository:
   ```
   git clone 'your-github-repo-link' # Here https://github.com/SMJNayeem/aws-vpc-controller.git is the repo link (basically empty repo)
   cd 'your-repo-name' # Here aws-vpc-controller is the repo name
   ```

2. Install dependencies:
   ```
   go mod init
   go mod tidy
   ```
3. After installing the dependencies, use kubebuilder to initialize the project:
   ```
   kubebuilder init --domain example.com --repo github.com/Username/repo-name   # Here kubebuilder init --domain poridhi.io --repo github.com/SMJNayeem/aws-vpc-controller
   ```

4. Create the API:
   ```
   kubebuilder create api --group vpc --version v1 --kind VPC
   ```

5. Modify the CRD files, Controller files and Main files:

    Edit the vpc_types.go file, vpc_controller.go file and main.go file as given in the repo
    - vpc_types.go file:
    ```
    type VPCSpec struct {
        // CIDR block for the VPC
        CIDR string `json:"cidr"`

        // Name of the VPC
        Name string `json:"name"`

        // Region where the VPC should be created
        Region string `json:"region"`
    }

    // VPCStatus defines the observed state of VPC
    type VPCStatus struct {
        // ID of the created VPC
        VPCID string `json:"vpcId,omitempty"`

        // Current state of the VPC
        State string `json:"state,omitempty"`

        // Any error message
        Error string `json:"error,omitempty"`
    }

    //+kubebuilder:object:root=true
    //+kubebuilder:subresource:status

    // VPC is the Schema for the vpcs API
    type VPC struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   VPCSpec   `json:"spec,omitempty"`
        Status VPCStatus `json:"status,omitempty"`
    }

    type VPCList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`
        Items           []VPC `json:"items"`
    }

    // Ensure VPCList implements runtime.Object
    func (v *VPCList) DeepCopyObject() runtime.Object {
        return v.DeepCopy() // Assuming DeepCopy is generated
    }

    func init() {
        SchemeBuilder.Register(&VPC{}, &VPCList{})
    }
    ```
    - internal/controller/vpc_controller.go:
    ```
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
    ```

    - main.go file:
    ```
    import (
        "crypto/tls"
        "flag"
        "os"

        // Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
        // to ensure that exec-entrypoint and run can make use of them.
        _ "k8s.io/client-go/plugin/pkg/client/auth"

        "k8s.io/apimachinery/pkg/runtime"
        utilruntime "k8s.io/apimachinery/pkg/util/runtime"
        clientgoscheme "k8s.io/client-go/kubernetes/scheme"
        ctrl "sigs.k8s.io/controller-runtime"
        "sigs.k8s.io/controller-runtime/pkg/healthz"
        "sigs.k8s.io/controller-runtime/pkg/log/zap"
        "sigs.k8s.io/controller-runtime/pkg/metrics/filters"
        metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
        "sigs.k8s.io/controller-runtime/pkg/webhook"

        vpcv1 "github.com/SMJNayeem/aws-vpc-controller/api/v1"
        "github.com/SMJNayeem/aws-vpc-controller/internal/controller"
        // +kubebuilder:scaffold:imports
    )
    // The rest of the code is almost same as the kubebuilder init command generated code

    ```


6. Generate the CRD and RBAC manifests:

    Run the following command to generate the code and manifest the Custom Resource Definition (CRD) and RBAC permissions:
   ```
   make manifests
   ```

7. Build and push the Docker image:
   ```
   docker build -t your-dockerhub-username/aws-vpc-controller:v1 .
   docker push your-dockerhub-username/aws-vpc-controller:v1
   ```

8. Create a Kind cluster:
   ```
   kind create cluster --name aws-vpc-controller
   ```

9. Before Deploying, update the go files and import packages
   ```
   go mod tidy
   ``` 

10. Deploy the controller to the cluster:
   ```
   make install
   make deploy IMG=<your-dockerhub-username>/aws-vpc-controller:v1
   ```

11. Create a Secret for AWS credentials:
   ```
   kubectl create secret generic aws-credentials \
     --from-literal=AWS_ACCESS_KEY_ID=<your-access-key> \
     --from-literal=AWS_SECRET_ACCESS_KEY=<your-secret-key>
   ```

12. Update the controller deployment to use the AWS credentials:
   Edit the `config/manager/manager.yaml` file and apply the changes:
   ```
   kubectl apply -f config/manager/manager.yaml
   ```

13. Verify the deployment:
   ```
   kubectl get deployments -n 'your namespace' # Here 'your namespace' is 'aws-vpc-controller'
   kubectl describe deployment 'your deployment name' # Here 'your deployment name' is 'aws-vpc-controller-controller-manager'
   kubectl get pods --all-namespaces
   kubectl get pods -n aws-vpc-controller-system # Here 'aws-vpc-controller-system' is the namespace
   kubectl describe pod 'your pod name' -n aws-vpc-controller-system # Here 'your pod name' is 'aws-vpc-controller-controller-manager-598fbf7786-b82tm'
   ```

14. Check the rbac permissions, crd, Makefile, go mod for a smooth deployment



## Usage

1. Create a VPC Custom Resource:
In the config/samples/vpc_v1_vpc.yaml file, update the vpc name, cidr and region to your desired values
   ```yaml
   apiVersion: vpc.poridhi.io/v1
   kind: VPC
   metadata:
     name: my-vpc
   spec:
     cidr: "10.0.0.0/16"
     name: "my-vpc"
     region: "ap-southeast-1"
   ```

2. Apply the CR:
   ```
   kubectl apply -f config/samples/vpc_v1_vpc.yaml
   ```

3. Check the status of your VPC:
   ```
   kubectl get vpc
   kubectl describe vpc my-vpc
   ```
   ![Image Describe](<images/Screenshot from 2024-10-14 23-44-34.png>)

4. To update the VPC, edit the CR and reapply:
   ```
   kubectl edit vpc my-vpc
   ```

5. To delete the VPC, delete the CR:
   ```
   kubectl delete vpc my-vpc
   ```
   ![Image Delete](<images/Screenshot from 2024-10-15 01-00-42.png>)

## Development

To add new features or modify the existing ones, follow these steps:

1. Make the changes in the code.
2. Run `make manifests` to update the CRD.
3. Make sure the Makefile, go mod and code generated by kubebuilder, rbac permissions, crd, controller and main files are correct
4. Run `make run` to run the controller locally.