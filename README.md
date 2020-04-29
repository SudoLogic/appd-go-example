# appd-go-example

### Instructions

This is a simple, working example of instrumenting a Go app 
with the [Appdynamics Go SDK](https://download.appdynamics.com/download/).

With the appdynamics SDK within your GOPATH, make sure you have the following environment 
variables set at the system level, or in your IDE, or just hardcode them in the books.go file:

APPD_CONT (Your controller URL) <br>
APPD_ACC  (Your APPD Account name (If saas, first segment of controller url)) <br>
APPD_KEY  (Your APPD Access Key)

That's all it should take. Run the app and then check appdynamics. Give it about 5 minutes of running and 
you should see a flow map, backends, and eventually snapshots, which should have data collectors attached.
