package cube

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/anuvu/cube/component"
	"github.com/anuvu/cube/config"
	. "github.com/smartystreets/goconvey/convey"
)

type tester struct {
	configError bool
}

func newtest(ctx component.Context) *tester {
	return &tester{}
}

func newBadConfig() *tester {
	return &tester{true}
}

func (d *tester) Config() config.Config {
	return nil
}

func (d *tester) Configure(ctx component.Context) error {
	if d.configError {
		return errors.New("bad config")
	}
	return nil
}

func (d *tester) Start(ctx component.Context) error {
	return errors.New("bad start")
}

type stoptester struct {
}

func (d *stoptester) Stop(ctx component.Context) error {
	return errors.New("bad stop")
}

func TestCubePanics(t *testing.T) {
	// Replace os.Args for test case
	oldArgs := os.Args
	os.Args = []string{"cube.test"}
	defer func() { os.Args = oldArgs }()

	Convey("cube main should panic on create error", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			err := g.Add(func(bool) int { return 0 })
			return nil, err
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})

	Convey("cube main should panic on config error", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			err := g.Add(newBadConfig)
			return nil, err
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})

	Convey("cube main should panic dependencies are not met", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			err := g.Add(func(i *int) {})
			return nil, err
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})

	Convey("cube main should panic on start errors", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			g.Add(newtest)
			return nil, nil
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})
	Convey("cube main should panic on invoke errors", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			return func() error {
				return errors.New("bad invoke")
			}, nil
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})

	Convey("cube main should panic on stop errors", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			g.Add(func() *stoptester { return &stoptester{} })
			g.Add(func(s *stoptester, k component.ServerShutdown) int { k(); return 0 })
			return nil, nil
		}
		So(func() { Main(initFunc) }, ShouldPanic)
	})

	Convey("calling shutdown handler should stop server", t, func() {
		initFunc := func(g component.Group) (Invoker, error) {
			g.Add(func(s *shutDownHandler) int {
				s.shut(syscall.SIGTERM)
				return 0
			})
			return nil, nil
		}
		So(func() { Main(initFunc) }, ShouldNotPanic)
	})
}
