package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	redisv1beta1 "redis/pkg/apis/qwoptcontroller/v1beta1"
)

// 调谐结构体，一版都要包含 client.Client 属性
type RedisReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// 将 Manager 与 Controller 关联
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr). // 创建 Controller
							For(&redisv1beta1.Redis{}). // 监听哪类资源
							Complete(r)                 // 调谐结构体(结构体必须实现了调谐接口- Reconcile)
}

// 调谐逻辑，Redis 资源的监控处理逻辑
// 注意，应该通过返回控制是否将本次请求重新入队
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("redis", req.NamespacedName)
	log.Info("reconciling redis")

	// 不重新入队列
	forget := ctrl.Result{
		Requeue:      false,
		RequeueAfter: 0,
	}

	// redis 实例
	rds := &redisv1beta1.Redis{}
	if err := r.Client.Get(ctx, req.NamespacedName, rds); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "unable to get redis")
			return forget, nil
		}
		return forget, err
	}

	log.Info("Add/Update/Delete for redis", "name", rds.GetName())

	return forget, nil
}