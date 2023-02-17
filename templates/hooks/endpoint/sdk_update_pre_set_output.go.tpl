
// always requeue with desired state and return immediately due to eventually
// consistent API
return desired, ackrequeue.NeededAfter(nil, defaultRequeueDelay)

// TODO (@embano1): we can't tell code-gen to not generate the rest of the code
