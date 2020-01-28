package log

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Logger interface {
	logr.InfoLogger

	Error(err error, msg string, keysAndValues ...interface{})

	V(level int) logr.InfoLogger

	WithValues(keysAndValues ...interface{}) Logger

	WithName(name string) Logger

	WithNamespace(namespace string) Logger

	WithProjectName(name client.ObjectKey) Logger

	WithClusterName(clusterName client.ObjectKey) Logger
}

func NewLogger(baseLogger logr.Logger) Logger {
	return &logger{baseLogger}
}

type logger struct {
	logr logr.Logger
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.logr.Info(msg, keysAndValues...)
}

func (l *logger) Enabled() bool {
	return l.logr.Enabled()
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logr.Error(err, msg, keysAndValues...)
}

func (l *logger) V(level int) logr.InfoLogger {
	return l.logr.V(level)
}

func (l *logger) WithNamespace(namespace string) Logger {
	return &logger{
		l.logr.WithValues("namespace", namespace),
	}
}

func (l *logger) WithClusterName(clusterName types.NamespacedName) Logger {
	return &logger{
		l.logr.WithValues("cluster", clusterName),
	}
}

func (l *logger) WithProjectName(projectName types.NamespacedName) Logger {
	return &logger{
		l.logr.WithValues("project", projectName),
	}
}

func (l *logger) WithValues(keysAndValues ...interface{}) Logger {
	return &logger{
		l.logr.WithValues(keysAndValues...),
	}
}

func (l *logger) WithName(name string) Logger {
	return &logger{
		l.logr.WithName(name),
	}
}
