package cube_test

import (
	"os"
	"time"

	"github.com/anuvu/cube"
	"github.com/anuvu/cube/component"
	"github.com/anuvu/cube/config"
)

type dummy struct {
}

func newDummy(ctx component.Context) *dummy {
	ctx.Log().Info().Msg("dummy object created")
	return &dummy{}
}

func (d *dummy) Configure(ctx component.Context, store config.Store) error {
	ctx.Log().Info().Msg("dummy object configured")
	return nil
}

func (d *dummy) Start(ctx component.Context) error {
	ctx.Log().Info().Msg("dummy object started")
	return nil
}

func (d *dummy) Stop(ctx component.Context) error {
	ctx.Log().Info().Msg("dummy object stopped")
	return nil
}

func killer(ctx component.Context, k component.ServerShutdown) {
	time.Sleep(time.Millisecond)
	ctx.Log().Info().Msg("Killing the server")
	k()
}

func ExampleMain() {
	// Replace os.Args for test case
	oldArgs := os.Args
	os.Args = []string{"cube.test"}
	defer func() { os.Args = oldArgs }()

	cube.Main(func(g component.Group) (cube.Invoker, error) {
		g.Add(newDummy)
		return func() error {
			g.Invoke(killer)
			return nil
		}, nil
	})

	// Output:
	// {"level":"info","name":"cube.test-core","message":"creating group"}
	// {"level":"info","name":"cube.test","message":"creating group"}
	// {"level":"info","name":"cube.test","message":"dummy object created"}
	// {"level":"info","name":"cube.test-core","message":"configuring group"}
	// {"level":"info","name":"cube.test","message":"configuring group"}
	// {"level":"info","name":"cube.test-core","message":"starting group"}
	// {"level":"info","name":"cube.test","message":"starting group"}
	// {"level":"info","name":"cube.test","message":"dummy object started"}
	// {"level":"info","name":"cube.test","message":"Killing the server"}
	// {"level":"info","name":"cube.test","message":"stopping group"}
	// {"level":"info","name":"cube.test","message":"dummy object stopped"}
	// {"level":"info","name":"cube.test-core","message":"stopping group"}
}
