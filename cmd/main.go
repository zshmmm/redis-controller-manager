package main

import (
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	redisv1beta1 "redis/pkg/apis/qwoptcontroller/v1beta1"
	"redis/controller"
)

const (
	Namespace = "default"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// 注册资源到 Schema
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(redisv1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctrl.SetLogger(zap.New())

	// 创建 Manger，并设置 Leader Election 相关的参数
	// ctrl.GetConfigOrDie() 自动的获取 kubeconfig，使用规则顺序如下：
	// 1. --kubeconfig flag pointing at a file
	// 2. KUBECONFIG environment variable pointing at a file
	// 3. In-cluster config if running in cluster
	// 4. $HOME/.kube/config if exists.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Namespace: Namespace, // 监听特定 namespace
		Scheme:    scheme,
		// 为了保证 Controller 的高可用，我们常常同时运行多个 Controller 实例。
		// 在这种情况下，多个 Controller 实例之间需要进行 Leader Election。
		// 被选中成为 Leader 的 Controller 实例才执行 Watch 和 Reconcile 逻辑，其余 Controller 处于等待状态。
		// Manager 实现了该功能，并能够自动创建 Lease 资源。
		// 另外 Manager 实现了更多功能 (Health, Metric, webhook)，在后续 kubebuilder 例子中会展示
		LeaderElection:   true,
		LeaderElectionID: "ecaf1259.redis.qwoptcontroller.k8s.io",
		// 当在 k8s 集群完运行时，需要定义选举资源所在的命名空间
		LeaderElectionNamespace: Namespace,
	})

	if err != nil {
		setupLog.Error(err, "start manager failed")
		os.Exit(1)
	}

	// 实例化 RedisReconciler ，并将其关联到 mgr
	if err = (&controller.RedisReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "create controller failed", "Controller", "Redis")
	}

	setupLog.Info("starting manager")

	// 启动 Manager，将启动起管理的所有 Controller
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "start manager failed")
		os.Exit(1)
	}
}
