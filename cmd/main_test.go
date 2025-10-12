package main

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type fakeManager struct {
	startErr error
	started  bool
}

func (f *fakeManager) Start(context.Context) error {
	f.started = true
	return f.startErr
}

func restoreGlobals() func() {
	originalExit := exitFunc
	originalGetConfig := getConfig
	originalNewManager := newManager
	originalSignal := signalHandler
	originalSetup := setupReconciler
	originalHealth := addHealthChecks

	return func() {
		exitFunc = originalExit
		getConfig = originalGetConfig
		newManager = originalNewManager
		signalHandler = originalSignal
		setupReconciler = originalSetup
		addHealthChecks = originalHealth
	}
}

func TestRunSuccess(t *testing.T) {
	defer restoreGlobals()()

	fake := &fakeManager{}
	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return fake, nil }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }
	setupReconciler = func(managerFacade) error { return nil }
	addHealthChecks = func(managerFacade, string) error { return nil }

	if code := run(nil); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !fake.started {
		t.Fatalf("expected Start to be called")
	}
}

func TestMainFunction(t *testing.T) {
	defer restoreGlobals()()

	fake := &fakeManager{}
	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return fake, nil }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }
	setupReconciler = func(managerFacade) error { return nil }
	addHealthChecks = func(managerFacade, string) error { return nil }

	var exitCode int
	exitFunc = func(code int) { exitCode = code }

	main()

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !fake.started {
		t.Fatalf("expected manager Start to be called")
	}
}

func TestRunManagerError(t *testing.T) {
	defer restoreGlobals()()

	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) {
		return nil, errors.New("boom")
	}
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }

	if code := run(nil); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunSetupError(t *testing.T) {
	defer restoreGlobals()()

	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return &fakeManager{}, nil }
	setupReconciler = func(managerFacade) error { return errors.New("fail") }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }

	if code := run(nil); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunHealthCheckError(t *testing.T) {
	defer restoreGlobals()()

	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return &fakeManager{}, nil }
	addHealthChecks = func(managerFacade, string) error { return errors.New("health") }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }
	setupReconciler = func(managerFacade) error { return nil }

	if code := run(nil); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunParseError(t *testing.T) {
	defer restoreGlobals()()

	if code := run([]string{"--unknown"}); code != 1 {
		t.Fatalf("expected parse failure exit code, got %d", code)
	}
}

func TestRunEnableHTTP2AndInsecureMetrics(t *testing.T) {
	defer restoreGlobals()()

	fake := &fakeManager{}
	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return fake, nil }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }
	setupReconciler = func(managerFacade) error { return nil }
	addHealthChecks = func(managerFacade, string) error { return nil }

	args := []string{"--enable-http2", "--metrics-secure=false"}
	if code := run(args); code != 0 {
		t.Fatalf("expected success, got code %d", code)
	}
	if !fake.started {
		t.Fatalf("expected manager started")
	}
}

func TestDisableHTTP2Proto(t *testing.T) {
	cfg := &tls.Config{}
	disableHTTP2Proto(cfg)
	if len(cfg.NextProtos) != 1 || cfg.NextProtos[0] != "http/1.1" {
		t.Fatalf("expected http/1.1 proto, got %v", cfg.NextProtos)
	}
}

func TestRunStartError(t *testing.T) {
	defer restoreGlobals()()

	fake := &fakeManager{startErr: errors.New("start")}
	newManager = func(*rest.Config, ctrl.Options) (managerFacade, error) { return fake, nil }
	getConfig = func() *rest.Config { return &rest.Config{} }
	signalHandler = func() context.Context { return context.Background() }
	setupReconciler = func(managerFacade) error { return nil }
	addHealthChecks = func(managerFacade, string) error { return nil }

	if code := run(nil); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
