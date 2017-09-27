Lambda function written in Go
=============================

This directory include an example of a lambda function written in Go,
and its usage with mless.

This example used [aws-lambda-go-shim](https://github.com/eawsy/aws-lambda-go-shim)
to wrap the Go binary in a python shim to be run in AWS Lambda (as well as in mless).

The example code receive notification about file creation in S3.
If the files are images it reduces resolution and move them to a different folder,
other wise it just move them.

## Quick usage

### Prerequisites

You need to setup an mless environment as described in the main README.md file.

### Compilation

Go is a compiled language, so the automatic synchronization of source code is not enough,
and the code need to built. The `build.mlab` script in the src sub-directory build the sources.
The first run of the script is a bit "slow" (on the order of a minute), because it may need
to download a compilation container image and compile everything from scratch.
Farther compilation would be much faster (3-4 seconds).

### Setup the proxy

Setting up the proxy is similar to the hash example:

* As the environment choose python 2.7
* Set in the environment `MLESS_SERVER` to be `http://<your servername>:8000`
* Configure a trigger on creating an object in some S3 bucket
  (you should probably limit this only to a specific folder).
* Make sure to configure the lambda function with a role that allow it to read/write the above bucket.
* Choose an output folder (in the same bucket),
  and configure its name in the environment variable `OUTPUT_FOLDER`

### Test

* Start mlessd (by running the `mlessd.mlab` script in the main folder).
* Upload an image to the specific folder (the code support only jpeg images)
* Note that the lambda code is run locally and the image is moved with reduced resolution to the specified output folder.

### Modify code

A simple modification is to enable the code to handle png images.
First try to upload a png image and watch how it is moved as is to the output folder.

Now modify the code: open `handler.go` and to the list of import add the following line:

```
_ "image/png"
```

(anywhere on the list would do, but just after `"image/jpeg"` is "right" place).

Now build the code (using `src/build.mlab`), and try again.

You should see that the file was converted to jpeg (and the resolution changed).

### Chaining function

If the output folder happen to be the input folder for one of the hash example you should see
that after moving the file, the hash code is called to compute the hash.
You can watch how several lambda function interact in your "local" environment.

## Source organization and build

Go is a compiled language, it is not enough to upload the sources they need to be compiled.
The way that the code is organized in this example (and the generally recommended way),
is to have three subdirectories of the function directory.

* `src` - containing the source file
* `build` - containing the built code - this is the directory configured in the `template.yaml` file
* `cache` - containing intermediate compilation results

The `build` and `cache` directories are *host only directories* in the sense that their content is
create and used on the mlab host and the mlab synchronization process does not synchronize them.
So normally you won't see them on the desktop.

This is achieved by the follwoing two lines in `.mlignore`:

```
/example/**/build/
/example/**/cache/
```

### Caching in `build.mlab`

The `build.mlab` include non-trivial handling of the caching.
The good news is that the script is general enough to handle any Go lambda function
that uses `aws-lambda-go-shim` - so you don't need to understand it in order to use it.

On the other hand the explanation of what we are doing is not that complicated.

Go compiler is quite fast, so even without any caching the results are pretty quick,
but go is pretty good in incremental compilation and as impatient guy, I want to use it.

Normally it is enough just to cache `$GOENV/pkg`. But `aws-lambda-go-shim` compile the code as a plugin.
The standard container image does not contain the standard library compiled for the plugin build mode.
It will compile it on demand and put it under `/usr/local/go/pkg`.
So if we cache only `$GOENV/pkg` this directory won't be cached, and go will be forced to recompile the
standard library every time you start the container, and then it would consider the `$GOENV/pkg` cache stale
since it is not based on the current compiled version of the standard library.

Therefore the build script need to cache both directories.
