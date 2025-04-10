package waiter

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/denvrdata/go-denvr/api/v1/servers/virtual"
)

type ActionFunc func(ctx context.Context, args ...any) (any, error)
type CheckFunc func(ctx context.Context, resp any) (bool, any, error)

type WaiterOptions struct {
	Timeout  time.Duration
	Interval time.Duration
}

type Waiter struct {
	Action  ActionFunc
	Check   CheckFunc
	Options WaiterOptions
}

func Wait(ctx context.Context, waiter Waiter, args ...any) (any, error) {
	start := time.Now()
	for {
		resp, err := waiter.Action(ctx, args...)
		if err != nil {
			return nil, err
		}

		passed, result, err := waiter.Check(ctx, resp)
		if err != nil {
			return nil, err
		}

		if passed {
			return result, nil
		}

		if time.Since(start) > waiter.Options.Timeout {
			return nil, fmt.Errorf("Waiting on check function timed out")
		}

		time.Sleep(waiter.Options.Interval)
	}
}

func NewWaiter(client *any, method string, options ...WaiterOptions) (*Waiter, error) {
	var opts WaiterOptions

	if len(options) > 0 {
		opts = options[0]
	}

	action, err := getAction(client, method)
	if err != nil {
		return nil, err
	}

	check, err := getCheckFunc(client, method)
	if err != nil {
		return nil, err
	}

	return &Waiter{
		Action:  action,
		Check:   check,
		Options: opts,
	}, nil
}

func getAction(client any, method string) (ActionFunc, error) {
	methodVal := reflect.ValueOf(client).MethodByName(method)
	if !methodVal.IsValid() {
		return nil, fmt.Errorf("Method %s not found", method)
	}

	action := func(ctx context.Context, args ...any) (any, error) {
		// Convert args to reflect.Value slice
		vals := make([]reflect.Value, len(args)+1)
		vals[0] = reflect.ValueOf(ctx)
		for i, arg := range args {
			vals[i+1] = reflect.ValueOf(arg)
		}

		// Call the method
		results := methodVal.Call(vals)

		// Handle the results
		if len(results) == 0 {
			return nil, nil
		}

		// Check for error
		if len(results) > 1 {
			errVal := results[len(results)-1]
			if !errVal.IsNil() {
				return nil, errVal.Interface().(error)
			}
		}

		// Return the first result
		result := results[0].Interface()
		return &result, nil
	}

	return action, nil
}

func getCheckFunc(client any, method string) (CheckFunc, error) {
	clientPath := reflect.TypeOf(client).PkgPath()
	parts := strings.Split(clientPath, ".")
	clientPkg := parts[len(parts)-1]

	if clientPkg == "virtual" {
		if slices.Contains([]string{"CreateServer", "StartServer"}, method) {
			return VmOnlineCheck(client), nil
		}
		if method == "StopServer" {
			return VmOfflineCheck(client), nil
		}
	}

	// TODO: Add applications once we've generated it.

	return nil, fmt.Errorf("Check function not found for method %s", method)
}

func VmOnlineCheck(client any) CheckFunc {
	return func(ctx context.Context, resp any) (bool, any, error) {
		details, ok := resp.(*virtual.VirtualServerDetailsItem)
		if !ok {
			return false, nil, fmt.Errorf("expected *virtual.VirtualServerDetailsItem, got %T", resp)
		}

		c := client.(*virtual.Client)
		status, err := c.GetServer(
			ctx,
			&virtual.GetServerParams{
				Id:        *details.Id,
				Namespace: *details.Namespace,
				Cluster:   *details.Cluster,
			})

		if err != nil {
			return false, nil, err
		}

		if *status.Status == "ONLINE" {
			return true, status, nil
		}

		return false, status, nil
	}
}

func VmOfflineCheck(client any) CheckFunc {
	return func(ctx context.Context, resp any) (bool, any, error) {
		details, ok := resp.(*virtual.VirtualServerDetailsItem)
		if !ok {
			return false, nil, fmt.Errorf("expected *virtual.VirtualServerDetailsItem, got %T", resp)
		}

		c := client.(*virtual.Client)
		status, err := c.GetServer(
			ctx,
			&virtual.GetServerParams{
				Id:        *details.Id,
				Namespace: *details.Namespace,
				Cluster:   *details.Cluster,
			})

		if err != nil {
			return false, nil, err
		}

		if *status.Status == "OFFLINE" {
			return true, status, nil
		}

		return false, status, nil
	}
}
