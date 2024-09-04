package main

import (
	"bytes"
	"context"
	"errors"
	"image/color"
	"os"
	"slices"
	"sync"

	_ "embed"

	"github.com/golang/geo/r3"
	goutils "go.viam.com/utils"
	"golang.org/x/exp/maps"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/slam"
	"go.viam.com/rdk/spatialmath"
)

var (
	//go:embed pcds/small.pcd
	smallPCDBytes []byte

	//go:embed pcds/large.pcd
	bigPCDBytes []byte
)

var imageTypes = map[string]bool{
	"jpeg": true,
	"png":  true,
}

var imageTypeOptions = maps.Keys(imageTypes)

var colors = map[string]color.NRGBA{
	"white": {R: 255, G: 255, B: 255, A: 255},
	"red":   {R: 255, A: 255},
	"green": {G: 255, A: 255},
	"blue":  {B: 255, A: 255},
}

var colorOptions = maps.Keys(colors)

func init() {
	slices.Sort(colorOptions)
	slices.Sort(imageTypeOptions)
}

var Model = resource.NewModel("ncs", "slam", "nickslam")
var (
	Reset = "\033[0m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
)

type fake struct {
	mu sync.Mutex
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable
	big    bool
	logger logging.Logger
}

type Config struct {
	Big bool `json:"big,omitempty"`
}

func (c *Config) Validate(path string) ([]string, error) {
	return nil, nil
}

func newSlam(
	ctx context.Context,
	deps resource.Dependencies,
	conf resource.Config,
	logger logging.Logger,
) (slam.Service, error) {
	c, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}
	named := conf.ResourceName().AsNamed()
	return &fake{
		Named:  named,
		big:    c.Big,
		logger: logger,
	}, nil
}

func (f *fake) Position(context.Context) (spatialmath.Pose, error) {
	return spatialmath.NewPose(r3.Vector{X: 255, Y: 255, Z: 0}, spatialmath.NewZeroOrientation()), nil
}

const chunkSizeBytes = 1 * 1024 * 1024

func (f *fake) PointCloudMap(context.Context, bool) (func() ([]byte, error), error) {
	var b *bytes.Reader
	if f.big {
		b = bytes.NewReader(bigPCDBytes)
	} else {
		b = bytes.NewReader(smallPCDBytes)
	}

	chunk := make([]byte, chunkSizeBytes)
	fun := func() ([]byte, error) {
		bytesRead, err := b.Read(chunk)
		if err != nil {
			return nil, err
		}
		return chunk[:bytesRead], err
	}

	return fun, nil
}

func (f *fake) InternalState(context.Context) (func() ([]byte, error), error) {
	return nil, errors.New("InternalState unimplemented")
}

func (f *fake) Properties(ctx context.Context) (slam.Properties, error) {
	return slam.Properties{}, nil
}

func (f *fake) DoCommand(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, boom := extra["boom"]
	if boom {
		f.logger.Info(Cyan + "Boom" + Reset)
		os.Exit(1)
	}
	return nil, nil
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) (err error) {
	resource.RegisterService(
		slam.API,
		Model,
		resource.Registration[slam.Service, *Config]{Constructor: newSlam})

	module, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}
	if err := module.AddModelFromRegistry(ctx, slam.API, Model); err != nil {
		return err
	}

	err = module.Start(ctx)
	defer module.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func main() {
	goutils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs(Model.String()))
}
