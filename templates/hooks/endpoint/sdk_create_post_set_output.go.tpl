if !endpointAvailable(&resource{ko}) {
	return &resource{ko}, requeueWaitWhileCreating
}
