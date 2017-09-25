Instrumenting a lambda function to proxy only part of the traffic
=================================================================

In this example, instead of replacing the whole lambda function with a proxy,
we instrument it so it can handle most of the traffic, while only specific events
are proxied to be handled locally.

## Instrumenting the lambda function

The same python code (`proxy.py`) that can be used as the whole proxy,
is used here as a library.

The main handler code include the

```python
proxyPattern = os.environ.get("PROXY_PATTERN")
if proxyPattern and re.match(proxyPattern,key):
    return proxy.proxy(event,context, funcName="hashProxy")
```

This code look for a pattern (regular expression) in the lambda function environment.
If it find that patter and it matches the function it call the proxy code.
Otherwise it continue to handle it inline.

For example sake, when it call the proxy, it call it with a different function name
(if the `hashProxy` parameter was not specified it would have used the same name).

## Trying the example

* Zip the two source files (`hash.py` and `proxy.py`) and upload it as the lambda function
* Create an S3 bucket/folder for the testing
* Create the lambda function
    * python 2.7 environment, handler is `hash.hashFile`
    * Upload the zip files

Now you can try the normal code, upload a file and the code would hash it (both md5 and sha1).

Now edit the environment:
* Add `MLESS_SERVER` to point to your mless server
* Add `PROXY_PATTERN` to contain a pattern that would be sent (for example `.*test.*`)

Modify the local code in some and upload two files, One that matches the patter and once that does not.

### Avoiding loops

In this example both the real lambda function and the local code include proxying code.
By default they also have the same configuration, so the local code may attempt to proxy to itself.

Future versions of mless (and the proxy code) may avoid this automatically,
however in the current proof of concept version, there is no such feature.

To avoid such loop, we override the `PROXY_PATTERN` environment variable in the local `template.yaml` file.
