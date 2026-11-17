package provider

import "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

const (
	operationTimeoutDescription = "Duration for the complete operation, including retries, for example `30s` or `5m`. Defaults to `2m`."
	storedTimeoutDescription    = "Includes retries and uses the duration saved in state (default `2m`, minimum provider timeout `10s`)."
)

func crudTimeoutOptions() timeouts.Opts {
	return resourceTimeoutOptions(timeouts.Opts{Create: true, Read: true, Update: true, Delete: true})
}

func resourceTimeoutOptions(opts timeouts.Opts) timeouts.Opts {
	opts.CreateDescription = operationTimeoutDescription
	opts.ReadDescription = storedTimeoutDescription
	opts.UpdateDescription = operationTimeoutDescription
	opts.DeleteDescription = storedTimeoutDescription

	return opts
}
