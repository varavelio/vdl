package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/varavelio/tinta"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugintypes"
)

// cjsPolyfill is required to allow plugins to use the CommonJS module system.
const cjsPolyfill = `
	var exports = {};
	var module = { exports: exports };
`

// scriptWrapper is a wrapper function that allows the plugin to be called with a
// string input. For the end user, the plugin should export a function called "generate"
// that takes an object as input and returns an object as output. The wrapper will handle
// the JSON parsing and stringifying to allow for easy communication between Go and JavaScript.
const scriptWrapper = `
	function __vdl__generate__find() {
		if (
			typeof exports === 'object' &&
			typeof exports.generate === 'function'
		) {
			return exports.generate;
		}
		if (
			typeof module === 'object' &&
			typeof module.exports === 'object' &&
			typeof module.exports.generate === 'function'
		) {
			return module.exports.generate;
		}
		if (typeof generate === 'function') {
			return generate;
		}
		throw new Error('the plugin does not export a "generate" function. Make sure to use "exports.generate = (input) => {...}"');
	}

	function __vdl__generate(inputString) {
		let genFn = __vdl__generate__find();
		const parsedInput = JSON.parse(inputString);
		const output = genFn(parsedInput);
		return JSON.stringify(output);
	}
`

// runPlugin executes the given plugin script with the provided input and returns the
// output or an error if one occurred.
func runPlugin(
	pluginName string,
	script string,
	input plugintypes.PluginInput,
) (plugintypes.PluginOutput, error) {
	// Create the JavaScript runtime using goja.
	vm := goja.New()

	// Inject a simple console polyfill to allow plugins to log messages.
	console := vm.NewObject()
	err := console.Set("log", func(call goja.FunctionCall) goja.Value {
		tinta.Text().Bold().Print(buildPluginLogPrefix(pluginName, "log"))
		for _, arg := range call.Arguments {
			fmt.Printf("%v ", arg.Export())
		}
		fmt.Println()
		return goja.Undefined()
	})
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	err = console.Set("error", func(call goja.FunctionCall) goja.Value {
		tinta.Text().Bold().Print(buildPluginLogPrefix(pluginName, "error"))
		for _, arg := range call.Arguments {
			fmt.Printf("%v ", arg.Export())
		}
		fmt.Println()
		return goja.Undefined()
	})
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	err = console.Set("warn", func(call goja.FunctionCall) goja.Value {
		tinta.Text().Yellow().Bold().Print(buildPluginLogPrefix(pluginName, "warn"))
		for _, arg := range call.Arguments {
			fmt.Printf("%v ", arg.Export())
		}
		fmt.Println()
		return goja.Undefined()
	})
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	err = console.Set("info", func(call goja.FunctionCall) goja.Value {
		tinta.Text().Cyan().Bold().Print(buildPluginLogPrefix(pluginName, "info"))
		for _, arg := range call.Arguments {
			fmt.Printf("%v ", arg.Export())
		}
		fmt.Println()
		return goja.Undefined()
	})
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	err = vm.Set("console", console)
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	// Inject the polyfill to allow plugins to use CJS module system (exports, module.exports).
	if _, err := vm.RunString(cjsPolyfill); err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing environment: %w", err)
	}

	// Run the plugin script in the VM.
	if _, err := vm.RunString(script); err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error evaluating the plugin: %w", err)
	}

	// Inject the wrapper function that will call the plugin's generate function. This wrapper
	// is responsible for parsing the input string, calling the plugin's generate function, and
	// stringifying the output.
	if _, err := vm.RunString(scriptWrapper); err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error initializing wrapper: %w", err)
	}

	// Get the wrapper function from the VM. This function is what we will call
	// to execute the plugin.
	wrapper, ok := goja.AssertFunction(vm.Get("__vdl__generate"))
	if !ok {
		return plugintypes.PluginOutput{}, fmt.Errorf("error finding the __vdl__generate wrapper function in the plugin script")
	}

	// Serialize the input from Go to a JSON string that can be passed to the plugin.
	// The plugin will then parse this string back into an object.
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("error serializing the input: %w", err)
	}

	// Call the wrapper function with the input string. The wrapper will call the
	// plugin's generate function and return the output as a JSON string.
	res, err := wrapper(goja.Undefined(), vm.ToValue(string(inputBytes)))
	if err != nil {
		// If plugin does a "throw new Error()" or the execution fails for any
		// reason, we catch it here and return an error.
		return plugintypes.PluginOutput{}, fmt.Errorf("the plugin failed during generation: %w", err)
	}

	// Deserialize the output from the plugin back into a Go struct. The plugin is expected
	// to return an object that matches the PluginOutput struct, which will be serialized
	// to JSON and then deserialized here.
	var output plugintypes.PluginOutput
	if err := json.Unmarshal([]byte(res.String()), &output); err != nil {
		return plugintypes.PluginOutput{}, fmt.Errorf("the plugin returned an invalid output: %w", err)
	}

	return output, nil
}

func buildPluginLogPrefix(pluginName string, prefix string) string {
	pluginName = strings.TrimSpace(pluginName)
	if pluginName == "" {
		return fmt.Sprintf("[%s] ", prefix)
	}
	return fmt.Sprintf("[%s %s] ", pluginName, prefix)
}
