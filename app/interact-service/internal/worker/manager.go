package worker

import "time"

var (
	resourceTypes = []string{"article"}
)

type Worker interface {
	Init()                        // 初始化任务信息
	DoWork()                      // 执行具体任务
	GetFrenquency() time.Duration // 获取任务执行频率
}

type Manager struct {
	workers []Worker
}

// AddWorker 添加一个 Worker 到管理中
func (m *Manager) AddWorker(worker Worker) {
	m.workers = append(m.workers, worker)
}

// Start 启动所有 Worker，根据频率定时执行任务
func (m *Manager) Start() {
	for _, worker := range m.workers {
		go func(w Worker) {
			w.Init()                                         // 初始化工作任务
			ticker := time.NewTicker(worker.GetFrenquency()) // 示例：每10秒执行一次
			for range ticker.C {
				w.DoWork()
			}
		}(worker)
	}
}
