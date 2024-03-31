package casbinbunadapter

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
)

func mockExecuteWithContextTimeOut(ctx context.Context, fn func() error) error {
	done := make(chan error)
	go func() {
		time.Sleep(500 * time.Microsecond)
		done <- fn()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func clearDBPolicy() (*casbin.Enforcer, *ctxBunAdapter) {
	ca, err := NewCtxAdapter("mysql", "root:root@tcp(127.0.0.1:3306)/test", WithDebugMode())
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer("examples/rbac_model.conf", ca)
	if err != nil {
		panic(err)
	}
	e.ClearPolicy()
	if err := e.SavePolicy(); err != nil {
		panic(err)
	}
	return e, ca
}

func TestCtxBunAdapter_AddPolicyCtx(t *testing.T) {
	e, ca := clearDBPolicy()

	if err := ca.AddPolicyCtx(context.Background(), "p", "p", []string{"alice", "data1", "read"}); err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}
	_ = e.LoadPolicy()
	testGetPolicy(
		t,
		e,
		[][]string{
			{"alice", "data1", "read"},
		},
	)

	var p = gomonkey.ApplyFunc(executeWithContext, mockExecuteWithContextTimeOut)
	defer p.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Microsecond)
	defer cancel()
	assert.EqualError(t, ca.AddPolicyCtx(ctx, "p", "p", []string{"alice", "data2", "read"}), "context deadline exceeded")
}

func TestCtxBunAdapter_RemovePolicyCtx(t *testing.T) {
	e, ca := clearDBPolicy()

	_ = ca.AddPolicyCtx(context.Background(), "p", "p", []string{"alice", "data1", "read"})
	_ = ca.AddPolicyCtx(context.Background(), "p", "p", []string{"alice", "data2", "read"})
	_ = ca.RemovePolicyCtx(context.Background(), "p", "p", []string{"alice", "data1", "read"})
	_ = e.LoadPolicy()
	testGetPolicy(
		t,
		e,
		[][]string{
			{"alice", "data2", "read"},
		},
	)

	var p = gomonkey.ApplyFunc(executeWithContext, mockExecuteWithContextTimeOut)
	defer p.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Microsecond)
	defer cancel()
	assert.EqualError(t, ca.RemovePolicyCtx(ctx, "p", "p", []string{"alice", "data2", "read"}), "context deadline exceeded")
}

func TestCtxBunAdapter_RemoveFilteredPolicyCtx(t *testing.T) {
	e, ca := clearDBPolicy()

	_ = ca.AddPolicyCtx(context.Background(), "p", "p", []string{"alice", "data1", "read"})
	_ = ca.AddPolicyCtx(context.Background(), "p", "p", []string{"alice", "data2", "read"})
	_ = ca.AddPolicyCtx(context.Background(), "p", "p", []string{"bob", "data1", "read"})
	_ = ca.RemoveFilteredPolicyCtx(context.Background(), "p", "p", 0, "alice")
	_ = e.LoadPolicy()
	testGetPolicy(
		t,
		e,
		[][]string{
			{"bob", "data1", "read"},
		},
	)

	var p = gomonkey.ApplyFunc(executeWithContext, mockExecuteWithContextTimeOut)
	defer p.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Microsecond)
	defer cancel()
	assert.EqualError(t, ca.RemoveFilteredPolicyCtx(ctx, "p", "p", 0, "alice"), "context deadline exceeded")
}
