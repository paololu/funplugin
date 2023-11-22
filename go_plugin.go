package funplugin

import (
	"fmt"
	"plugin"
	"reflect"
	"runtime"

	"github.com/paololu/funplugin/fungo"
)

// goPlugin implements golang official plugin
type goPlugin struct {
	*plugin.Plugin
	path            string                   // plugin file path
	cachedFunctions map[string]reflect.Value // cache loaded functions to improve performance
}

func newGoPlugin(path string) (*goPlugin, error) {
	if runtime.GOOS == "windows" {
		logger.Warn("go plugin does not support windows")
		return nil, fmt.Errorf("go plugin does not support windows")
	}

	// logger
	logger = logger.ResetNamed("go-plugin")

	plg, err := plugin.Open(path)
	if err != nil {
		logger.Error("load go plugin failed", "path", path, "error", err)
		return nil, err
	}

	logger.Info("load go plugin success", "path", path)
	p := &goPlugin{
		Plugin:          plg,
		path:            path,
		cachedFunctions: make(map[string]reflect.Value),
	}
	return p, nil
}

func (p *goPlugin) Type() string {
	return "go-plugin"
}

func (p *goPlugin) Path() string {
	return p.path
}

func (p *goPlugin) Has(funcName string) bool {
	logger.Debug("check if plugin has function", "funcName", funcName)
	fn, ok := p.cachedFunctions[funcName]
	if ok {
		return fn.IsValid()
	}

	sym, err := p.Plugin.Lookup(funcName)
	if err != nil {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}
	fn = reflect.ValueOf(sym)

	// check function type
	if fn.Kind() != reflect.Func {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}

	p.cachedFunctions[funcName] = fn
	return true
}

func (p *goPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	if !p.Has(funcName) {
		return nil, fmt.Errorf("function %s not found", funcName)
	}
	fn := p.cachedFunctions[funcName]
	return fungo.CallFunc(fn, args...)
}

func (p *goPlugin) Quit() error {
	// no need to quit for go plugin
	return nil
}
