# AWS VPC কন্ট্রোলার

এই প্রোজেক্টটি একটি AWS VPC কন্ট্রোলার তৈরি করে যা Kubernetes অপারেটর প্যাটার্ন ব্যবহার করে। এটির কাজ হলো Kubernetes Custom Resource (CR) এর মাধ্যমে সরাসরি AWS VPC তৈরি, আপডেট এবং মুছে ফেলা। কিভাবে Kubernetes, কন্ট্রোলার, কাস্টম রিসোর্স ডেফিনিশন (CRD), এবং অন্যান্য কারিগরি ধারণা এই প্রকল্পে ভূমিকা পালন করে, তা বিস্তারিতভাবে আলোচনা করা হলো।

## সূচিপত্র

1. [বেসিক ধারণা](#বেসিক-ধারণা)
2. [আর্কিটেকচার ওভারভিউ](#আর্কিটেকচার-ওভারভিউ)
3. [সিস্টেম ফ্লো এবং Dependency](#সিস্টেম-ফ্লো-এবং-Dependency)
4. [পূর্বশর্ত](#পূর্বশর্ত)
5. [সেটআপ এবং ইনস্টলেশন](#সেটআপ-এবং-ইনস্টলেশন)
6. [ব্যবহার](#ব্যবহার)
7. [ডেভেলপমেন্ট](#ডেভেলপমেন্ট)


## বেসিক ধারণা

- **Kubernetes এবং এর ভূমিকা**

Kubernetes একটি ওপেন-সোর্স কন্টেইনার অর্কেস্ট্রেশন টুল যা কন্টেইনারাইজড অ্যাপ্লিকেশনগুলির ডিপ্লয়মেন্ট, স্কেলিং এবং ম্যানেজমেন্ট অটোমেট করে। এটি ক্লাস্টার নামে পরিচিত একটি সিস্টেমে একাধিক "নোড" বা কন্টেইনার হোস্ট করে এবং বিভিন্ন অ্যাপ্লিকেশন বা মাইক্রোসার্ভিস চালায়। ক্লাস্টারের মধ্যে Pods, Services, Deployments ইত্যাদি Kubernetes অবজেক্ট হিসেবে পরিচিত।

- **কন্ট্রোলার এবং কন্ট্রোল লুপ**

Kubernetes এ, কন্ট্রোলার হল একটি কন্ট্রোল লুপ যা নির্দিষ্ট Kubernetes অবজেক্টের অবস্থা পর্যবেক্ষণ করে এবং প্রয়োজনে পরিবর্তন আনয়ন করে। উদাহরণস্বরূপ, একটি Deployment কন্ট্রোলার নিশ্চিত করে যে নির্দিষ্ট সংখ্যক Pod চলমান আছে। কন্ট্রোলারের কাজ হলো ডেসায়ার্ড স্টেট এবং কারেন্ট স্টেট এর মধ্যে সামঞ্জস্য বজায় রাখা, যাতে অ্যাপ্লিকেশন সর্বদা নির্দিষ্ট অবস্থা বজায় রাখতে পারে। কন্ট্রোলারগুলি মূলত Golang-এ লেখা এবং বিভিন্ন API এর মাধ্যমে Kubernetes API সার্ভারের সাথে যোগাযোগ করে।

- **Custom Resource এবং Custom Resource Definition (CRD)**

Kubernetes শুধুমাত্র বিল্ট-ইন অবজেক্ট যেমন Pods, Services ইত্যাদির মধ্যেই সীমাবদ্ধ নয়। Kubernetes Custom Resources (CR) ব্যবহারকারীদের কাস্টম অবজেক্ট তৈরি করতে দেয়, যা Kubernetes ক্লাস্টারে সংরক্ষণ করা যায় এবং ব্যবহৃত হতে পারে। Custom Resource Definition (CRD) ব্যবহার করে Custom Resources এর প্রকারভেদ সংজ্ঞায়িত করা হয়। CRDs মূলত API সার্ভারে কাস্টম স্কিমা যোগ করে এবং এটি Kubernetes ক্লাস্টারে নতুন ধরনের অবজেক্ট ব্যবহারের সুযোগ দেয়। এই প্রকল্পে, VPC Custom Resource হল একটি CRD যা AWS VPC রিসোর্স কন্ট্রোল করতে ব্যবহৃত হয়।

- **Kubebuilder এবং Operator প্যাটার্ন**

Kubebuilder একটি ফ্রেমওয়ার্ক যা Kubernetes APIs নির্মাণ ও ব্যবস্থাপনার জন্য ব্যবহৃত হয়। এটি কাস্টম রিসোর্স ডেফিনিশনস (CRDs) তৈরি এবং কন্ট্রোলার নির্মাণকে সহজ করে তোলে। Operator প্যাটার্ন ব্যবহার করে Kubebuilder একটি অপারেটর নির্মাণ করে, যা মূলত একটি বিশেষ ধরনের কন্ট্রোলার যা কাস্টম রিসোর্সগুলি পরিচালনা করতে সক্ষম।

অপারেটর প্যাটার্নের মাধ্যমে Kubernetes ক্লাস্টারে কাস্টম লজিক যোগ করা যায় যা স্বয়ংক্রিয়ভাবে কাস্টম রিসোর্স পরিচালনা করতে সক্ষম। এই কন্ট্রোলার, বা অপারেটর, Kubernetes এর বিভিন্ন ইভেন্ট পর্যবেক্ষণ করে এবং সংশ্লিষ্ট কাজ সম্পাদন করে। এই ক্ষেত্রে, VPC কন্ট্রোলার হলো একটি অপারেটর যা VPC Custom Resource এর ইভেন্টগুলি পর্যবেক্ষণ করে এবং AWS VPC রিসোর্স পরিচালনা করে।

- **AWS VPC এবং এর সাথে ইন্টারঅ্যাকশন**

Amazon Virtual Private Cloud (VPC) হলো একটি ভার্চুয়াল নেটওয়ার্ক যা AWS অ্যাকাউন্টে প্রাইভেট ক্লাউড এনভায়রনমেন্ট তৈরি করে। VPC এর মাধ্যমে ব্যবহারকারীরা সাবনেট, রাউটিং টেবিল, ইন্টারনেট গেটওয়ে ইত্যাদি নির্মাণ করতে পারে এবং কাস্টমাইজড নেটওয়ার্ক পরিবেশ প্রস্তুত করতে পারে। AWS SDK ব্যবহার করে Kubernetes VPC কন্ট্রোলার সরাসরি AWS এর সাথে যোগাযোগ করে এবং প্রয়োজনীয় API কলের মাধ্যমে VPC তৈরি, আপডেট এবং মুছে ফেলার কাজ করে।

## আর্কিটেকচার ওভারভিউ: কিভাবে সিস্টেম কাজ করে

![Architecture Overview](<images/Screenshot from 2024-10-15 00-30-02.png>)

প্রোজেক্টটি একটি সাধারণ Kubernetes ক্লাস্টারে VPC কন্ট্রোলার স্থাপন করে, যা AWS VPC কাস্টম রিসোর্সগুলিকে পর্যবেক্ষণ করে এবং প্রয়োজনীয় কাজ করে। নিম্নলিখিত ধাপগুলি একটি সাধারণ ওভারভিউ প্রদান করে:

**VPC কন্ট্রোলার চালানো**: VPC কন্ট্রোলারটি Kubernetes ক্লাস্টারের একটি নোডে চালানো হয়, যেখানে এটি কাস্টম রিসোর্স ইভেন্টগুলি পর্যবেক্ষণ করে।

**Custom Resource পরিবর্তন পর্যবেক্ষণ**: Kubernetes API সার্ভার এর মাধ্যমে VPC Custom Resource তৈরি বা আপডেট করা হলে কন্ট্রোলারটি পরিবর্তনগুলি শনাক্ত করে।

**AWS SDK এর মাধ্যমে মিথস্ক্রিয়া**: পরিবর্তন সনাক্ত করার পর, VPC কন্ট্রোলার AWS SDK ব্যবহার করে প্রয়োজনীয় API কল করে, যা VPC তৈরি, আপডেট, বা মুছে ফেলে।

**VPC Custom Resource স্ট্যাটাস আপডেট**: VPC কন্ট্রোলার প্রাপ্ত তথ্য অনুযায়ী VPC Custom Resource এর স্ট্যাটাস আপডেট করে, যাতে ব্যবহারকারীরা এর বর্তমান অবস্থা দেখতে পারেন।


## সিস্টেম ফ্লো: কিভাবে কাজের ধাপগুলি পরিচালিত হয়

সিস্টেম ফ্লো নিম্নরূপ:

![System Flow and Dependencies](<images/Screenshot from 2024-10-15 00-30-21.png>)

1. ব্যবহারকারী kubectl এর মাধ্যমে VPC Custom Resource তৈরি/আপডেট করে: যখন ব্যবহারকারী kubectl এর মাধ্যমে একটি VPC Custom Resource (CR) প্রয়োগ করে, তখন এটি Kubernetes API সার্ভার দ্বারা গ্রহণ করা হয়।

2. Kubernetes API সার্ভার VPC কন্ট্রোলারকে অবহিত করে: API সার্ভার VPC কন্ট্রোলারকে CR পরিবর্তন সম্পর্কে অবহিত করে।

3. VPC কন্ট্রোলার AWS SDK ব্যবহার করে VPC তৈরি/আপডেট করে: কন্ট্রোলার প্রয়োজনীয় পরিবর্তনগুলি প্রয়োগ করে এবং AWS SDK ব্যবহার করে AWS VPC তৈরি বা আপডেটের জন্য API কল করে।

4. AWS VPC কন্ট্রোলারের কাছে তথ্য পাঠায়: যখন VPC তৈরি বা আপডেট সম্পন্ন হয়, তখন AWS থেকে প্রাপ্ত তথ্য VPC কন্ট্রোলারের কাছে পাঠানো হয়।

5. VPC কন্ট্রোলার VPC CR এর স্ট্যাটাস আপডেট করে: ফলাফল অনুযায়ী কন্ট্রোলার VPC Custom Resource এর স্ট্যাটাস আপডেট করে এবং API সার্ভারকে জানায়।

6. ব্যবহারকারী kubectl এর মাধ্যমে আপডেটেড স্ট্যাটাস দেখতে পায়: শেষে, ব্যবহারকারী kubectl কমান্ডের মাধ্যমে VPC Custom Resource এর স্ট্যাটাস দেখতে পারে।


## Kubernetes, CRDs এবং কন্ট্রোলার: ভিতরকার বিশ্লেষণ

Kubernetes কাস্টমাইজড অবজেক্ট পরিচালনার জন্য অত্যন্ত দক্ষ এবং সিস্টেমটি CRDs এবং কন্ট্রোলারগুলির মাধ্যমে ডিপ্লয়মেন্ট স্বয়ংক্রিয়করণে সক্ষম। CRD ব্যবহার করে Kubernetes ক্লাস্টারে নতুন ধরনের অবজেক্ট সংজ্ঞায়িত করা যায়, যা অ্যাপ্লিকেশনগুলির নির্দিষ্ট প্রয়োজনীয়তাগুলি পূরণ করতে পারে।

**কন্ট্রোলারের লজিক এবং ইভেন্ট হ্যান্ডলিং**

একটি কন্ট্রোলারের কাজ হলো Kubernetes API সার্ভার থেকে ইভেন্ট সংগ্রহ করা এবং নির্দিষ্ট লজিক অনুযায়ী কাজ করা। Kubernetes কন্ট্রোলার সবসময় "রিঅ্যাক্টিভ" থাকে, যার মানে এটি কেবল ইভেন্ট পরিবর্তন হলে কাজ শুরু করে। উদাহরণস্বরূপ, যখন একটি VPC Custom Resource আপডেট হয়, তখন কন্ট্রোলার AWS SDK ব্যবহার করে প্রয়োজনীয় কাজ করে।

**Kubernetes এবং Operator প্যাটার্ন এই প্রকল্পের জন্য আদর্শ কারণ:**

- স্বয়ংক্রিয় প্রভিশনিং: Kubernetes VPC কন্ট্রোলার কাস্টম রিসোর্স অনুযায়ী স্বয়ংক্রিয়ভাবে VPC তৈরি করতে পারে।
- ক্লাস্টার-লেভেল স্কেলিং: Kubernetes কন্ট্রোলার নিজেই স্কেল করতে সক্ষম এবং VPC সংক্রান্ত বিভিন্ন কাজগুলো সম্পাদন করতে পারে।
- বহুমুখী ব্যবহার: কাস্টমাইজড কন্ট্রোলার ব্যবহার করে বিভিন্ন ধরনের AWS সার্ভিসগুলো একত্রিত করা যায়।

এইভাবে, Kubernetes, CRDs, কন্ট্রোলার, এবং অপারেটর প্যাটার্ন মিলে এই সিস্টেমকে AWS VPC পরিচালনার জন্য একটি শক্তিশালী এবং স্কেলেবল প্ল্যাটফর্মে রূপান্তরিত করে।



## পূর্বশর্ত

- Go 1.20+
- Docker
- kubectl
- Kind
- AWS CLI যথাযথ ক্রেডেন্সিয়ালসহ কনফিগার করা
- Kubebuilder

## সেটআপ এবং ইনস্টলেশন

1. রিপোজিটরি ক্লোন করুন:
   ```
   git clone 'your-github-repo-link' # এখানে https://github.com/SMJNayeem/aws-vpc-controller.git হল রিপো লিংক (মূলত খালি রিপো)
   cd 'your-repo-name' # এখানে aws-vpc-controller হল রিপো নাম
   ```

2. Dependency ইনস্টল করুন:
   ```
   go mod init
   go mod tidy
   ```
3. Dependency ইনস্টল করার পরে, kubebuilder ব্যবহার করে প্রোজেক্টটি ইনিশিয়ালাইজ করুন:
   ```
   kubebuilder init --domain example.com --repo github.com/Username/repo-name   # এখানে kubebuilder init --domain poridhi.io --repo github.com/SMJNayeem/aws-vpc-controller
   ```

4. API তৈরি করুন:
   ```
   kubebuilder create api --group vpc --version v1 --kind VPC
   ```

5. CRD ফাইলগুলি, কন্ট্রোলার ফাইলগুলি এবং মেইন ফাইলগুলি পরিবর্তন করুন:

    vpc_types.go ফাইল, vpc_controller.go ফাইল এবং main.go ফাইল রিপোতে দেওয়া অনুযায়ী ইডিট করুন
    - vpc_types.go ফাইল:
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

    - main.go ফাইল:    
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

6. CRD এবং RBAC ম্যানিফেস্ট জেনারেট করুন:

    কাস্টম রিসোর্স ডেফিনিশন (CRD) এবং RBAC অনুমতিগুলি জেনারেট করতে নিম্নলিখিত কমান্ড চালান:   
    ```
   make manifests   
   ```

7. Docker ইমেজ বিল্ড এবং পুশ করুন:   
    ```
   docker build -t your-dockerhub-username/aws-vpc-controller:v1 .
   docker push your-dockerhub-username/aws-vpc-controller:v1   
   ```

8. Kind ক্লাস্টার তৈরি করুন:   
    ```
   kind create cluster --name aws-vpc-controller   
   ```

9. ডিপ্লয় করার আগে, go ফাইলগুলি আপডেট করুন এবং প্যাকেজগুলি ইমপোর্ট করুন   
    ```
   go mod tidy   
   ``` 

10. কন্ট্রোলারটি ক্লাস্টারে ডিপ্লয় করুন:
    ```
    make install
    make deploy IMG=<your-dockerhub-username>/aws-vpc-controller:v1   
    ```

11. AWS ক্রেডেন্সিয়ালগুলির জন্য একটি সিক্রেট তৈরি করুন:   
    ```
    kubectl create secret generic aws-credentials \
        --from-literal=AWS_ACCESS_KEY_ID=<your-access-key> \
        --from-literal=AWS_SECRET_ACCESS_KEY=<your-secret-key>   
     ```

12. কন্ট্রোলার ডিপ্লয়মেন্টটি AWS ক্রেডেন্সিয়ালগুলি ব্যবহার করতে আপডেট করুন:
   `config/manager/manager.yaml` ফাইলটি ইডিট করুন এবং পরিবর্তনগুলি প্রয়োগ করুন:   
    ```
    kubectl apply -f config/manager/manager.yaml   
    ```

13. ডিপ্লয়মেন্ট যাচাই করুন:   
    ```
    kubectl get deployments -n 'your namespace' # এখানে 'your namespace' হল 'aws-vpc-controller'
    kubectl describe deployment 'your deployment name' # এখানে 'your deployment name' হল 'aws-vpc-controller-controller-manager'
    kubectl get pods --all-namespaces
    kubectl get pods -n aws-vpc-controller-system # এখানে 'aws-vpc-controller-system' হল নেমস্পেস
    kubectl describe pod 'your pod name' -n aws-vpc-controller-system # এখানে 'your pod name' হল 'aws-vpc-controller-controller-manager-598fbf7786-b82tm'   
    ```

14. rbac অনুমতিগুলি, crd, Makefile, go mod এর জন্য যাচাই করুন যাতে সহজে ডিপ্লয় করা যায়



## ব্যবহার

1. একটি VPC Custom Resource তৈরি করুন:
config/samples/vpc_v1_vpc.yaml ফাইলে, vpc নাম, cidr এবং region আপনার পছন্দসই মানগুলি দিয়ে আপডেট করুন   
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

2. CR প্রয়োগ করুন:   
   ```
   kubectl apply -f config/samples/vpc_v1_vpc.yaml
   ```

3. আপনার VPC এর স্ট্যাটাস যাচাই করুন:   
   ```
   kubectl get vpc
   kubectl describe vpc my-vpc
   ```

   ![Image Describe](<images/Screenshot from 2024-10-14 23-44-34.png>)

4. VPC আপডেট করতে, CR ইডিট করুন এবং পুনরায় প্রয়োগ করুন:   
   ```
   kubectl edit vpc my-vpc
   ```

5. VPC মুছে ফেলতে, CR মুছে ফেলুন:   
   ```
   kubectl delete vpc my-vpc
   ```

   ![Image Delete](<images/Screenshot from 2024-10-15 01-00-42.png>)

## ডেভেলপমেন্ট

নতুন বৈশিষ্ট্য যোগ করতে বা বিদ্যমানগুলি পরিবর্তন করতে, নিম্নলিখিত পদক্ষেপগুলি অনুসরণ করুন:

1. কোডে পরিবর্তনগুলি করুন।
2. CRD আপডেট করতে `make manifests` চালান।
3. Makefile, go mod এবং kubebuilder দ্বারা জেনারেট করা কোড, rbac অনুমতিগুলি, crd, কন্ট্রোলার এবং মেইন ফাইলগুলি সঠিক আছে কিনা তা নিশ্চিত করুন।
4. কন্ট্রোলারটি লোকালি চালানোর জন্য `make run` চালান।