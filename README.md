# AWS VPC Controller

This project builds an AWS VPC controller using the Kubernetes Operator pattern. The controller's primary function is to create, update, and delete AWS VPCs directly through Kubernetes Custom Resources (CRs). This documentation provides a detailed overview of how Kubernetes, controllers, Custom Resource Definitions (CRDs), and other technical concepts play a role in the project.

## Table of Contents

1. [Basic Terms](#basic-terms)
2. [Architecture Overview](#architecture-overview)
3. [System Flow and Dependencies](#system-flow-and-dependencies)
4. [Prerequisites](#prerequisites)
5. [Setup and Installation](#setup-and-installation)
6. [Usage](#usage)
7. [Development](#development)

## Basic Terms

- **Kubernetes and Its Role**: Kubernetes is an open-source container orchestration tool that automates the deployment, scaling, and management of containerized applications. It manages multiple "nodes" (container hosts) within a system known as a cluster, running various applications or microservices. Inside a cluster, objects such as Pods, Services, and Deployments are created and managed.

- **Controller and Control Loop**: In Kubernetes, a controller is a control loop that continuously monitors the state of specific Kubernetes objects and makes changes to ensure the desired state is maintained. For example, a Deployment controller ensures a certain number of Pods are always running. Controllers, typically written in Golang, interact with the Kubernetes API server to manage these objects through API calls.

- **Custom Resources (CR) and Custom Resource Definitions (CRDs)**: Kubernetes extends beyond built-in objects like Pods and Services by allowing users to define custom objects using Custom Resources (CRs). A Custom Resource Definition (CRD) defines the schema and API for these custom objects. CRDs add new types to the Kubernetes API, enabling users to create and manage non-standard resources. In this project, the VPC Custom Resource is a CRD used to manage AWS VPCs directly from Kubernetes.

- **Kubebuilder and the Operator Pattern**: Kubebuilder is a framework for building and managing Kubernetes APIs. It simplifies the process of creating CRDs and controllers. Using the Operator pattern, Kubebuilder creates an operator, a specialized controller that manages custom resources.

    The Operator pattern allows the inclusion of custom logic within Kubernetes clusters, automating the management of complex resources. In this project, the VPC controller acts as an operator that monitors VPC Custom Resources and manages AWS VPCs accordingly.

- ***AWS VPC and Interaction with It**: Amazon Virtual Private Cloud (VPC) is a virtual network that provides a private cloud environment within an AWS account. With VPCs, users can configure subnets, routing tables, internet gateways, and more. The VPC controller interacts with AWS services through the AWS SDK, making API calls to create, update, or delete VPCs based on the defined CRs.


## Architecture Overview

![Architecture Overview](<images/Screenshot from 2024-10-15 00-30-02.png>)

The project deploys the VPC controller within a Kubernetes cluster. It watches over AWS VPC Custom Resources and performs necessary actions. Below is a high-level overview of the process:

**Running the VPC Controller**:
The VPC controller runs on a node within the Kubernetes cluster, watching for events related to VPC Custom Resources.

**Monitoring Custom Resource Changes**:
Whenever a VPC Custom Resource is created or updated, the Kubernetes API server notifies the VPC controller about the changes.

**Interaction with AWS SDK**:
After detecting changes, the controller uses the AWS SDK to make the appropriate API calls to create, update, or delete VPCs.

**Updating the Custom Resource Status**:
The controller updates the status of the VPC Custom Resource with the current state of the AWS VPC, providing visibility to users.

## System Flow and Dependencies

![System Flow and Dependencies](<images/Screenshot from 2024-10-15 00-30-21.png>)

1. User Creates or Updates a VPC CR using kubectl:
The user applies a VPC Custom Resource (CR) using kubectl. This action sends the resource information to the Kubernetes API server.

2. API Server Notifies the VPC Controller:
The API server triggers the VPC controller when a change in the VPC Custom Resource is detected.

3. VPC Controller Executes the Task via AWS SDK:
The controller processes the event and makes the necessary AWS API calls to create, update, or delete a VPC.

4. AWS Responds with Status Information:
AWS sends back the VPC creation or update status, which is captured by the controller.

5.  Controller Updates the VPC CR's Status:
Based on the AWS response, the controller updates the status of the VPC Custom Resource in Kubernetes.

6. User Checks Updated Status via kubectl:
The user can now view the updated status of the VPC Custom Resource using kubectl.


## Kubernetes, CRDs, and Controllers: A Deeper Analysis

Kubernetes efficiently manages customized objects, and the system leverages CRDs and controllers to automate deployments. Using a CRD, new types of objects can be defined to meet the specific needs of the application.

**Controller Logic and Event Handling**:
A controller's primary role is to react to events captured from the Kubernetes API server and apply the appropriate logic. In Kubernetes, controllers follow a reactive pattern, only executing tasks when an event occurs. For example, when a VPC CR is updated, the controller uses the AWS SDK to make necessary changes to the VPC.

**Why the Operator Pattern is Ideal for This Project**:

- Automated Provisioning: The VPC controller can automatically provision AWS VPCs based on the defined custom resources.

- Cluster-Level Scaling: The controller can scale itself within the Kubernetes cluster and handle multiple VPC operations efficiently.

- Versatility: With a custom controller, multiple AWS services can be integrated into the Kubernetes environment.

Thus, the combination of Kubernetes, CRDs, controllers, and the Operator pattern transforms the system into a robust, scalable platform for managing AWS VPCs efficiently.

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