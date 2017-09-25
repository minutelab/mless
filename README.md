Mless - Run AWS Lambda triggered events locally
===============================================

Mless enable running Lambda function in a "local" environment.
With mless the functions are triggered in the usual AWS lambda fashion like:

 * Call from API gateway
 * Any AWS Lambda supported event source such as
   * Changes to S3 bucket
   * Kinesis Streams events

When running the function will use the role configured for the lambda function,
and its configured environment. But the lambda function will run in the "local" environment:

 * Using the latest code with no need for deployment
   * For scripted languages (python, node) just save the code in the IDE
 * Enabling debugging with the IDE

 Though the code will look and fill as it is running locally, it will actually run in AWS,
 so it can be access private VPC resources.

## Proof of Concept

The code is currently in a proof of concept stage. Which basically mean:

 * Only python 2.7 environment is supported
 * There is no authentication/security for passing data
   between the lambda function proxy and the actual running environent
 * We are eager to hear feedback
 * Expect frequent updates

## Technology

Mless combine several technologies

* Mless uses the basic [MinuteLab](http://minutelab.io) technology to provide the look & feel of local environment for code that run in the cloud.
* Mless uses a small proxy inside the lambda function to transfer control to the local environment.
  This proxy can be used in two ways:
   * The proxy can be the whole lambda function and in this case it transfer all triggered events to the local environment
   * The proxy can be used as library inside the real lambda function to transfer only part of the events to the local environemnt

## MLess vs Sam Local

There are great similarities between mless and sam local.
Both projects aim to run lambda function code in a local environment.

In fact mless share some of the code with sam local.

The biggest difference is that while sam local run the function locally they are also triggered locally,
and run under the local user AWS role.

If you have a function that should react to a file written to S3 bucket, sam local would enable you to simulate such event and run the code.

How ever sam local won't allow you to react to real events. Or to view a chain of events. Suppose one lambda function should react to S3 bucket by writing something (sometime) to a Kinesis stream and another function should react based on this.

Sam local would enable you to simulate both events independently, but you will have to orchestrate them.
With mless once you start the chain of event, it will roll out as it would in production.

Another difference is that (at least currently) sam local run each function in its own container, it never reuses the same to container to run several function.
Therefore it does not allow to test the side effects of container reuses. Both the positive effects (like preparing cache) and the negative one.

## Usage

### Prerequisites

#### Setting up Minute Lab

Mless is built on top [MinuteLab](http://minutelab.io), so you will have to register and install the client.
The [Quick Start Guide](http://docs.minutelab.io/user-guide/quickstart/) is a good starting point. However for the purpose of mless you will need to set up an independent domain (currently only independent domain allow setting of security groups allowing inbound access).

Follow this guide to learn how to setup a domain, a host and a synced share between your desktop and the Minute Lab Host.

#### Setting inbound access

The lambda proxy would have to access a the mless container running on the host. For this you will have to configure a security group that allow inbound access to the port (by default 8000).

You will have to setup the security group in the EC2 console, you can attach it to the host either in the EC2 console or the MinuteLab console.

#### Setting dynamic DNS

To access the container the proxy code would have to know its IP. The easiest way is through some dynamic DNS service. You can use anything you like, but the mless code contain scripts template to use [DYNU](https://www.dynu.com).

Register to such a service to obtain a dns name (and credentials that allow you to register to it)

### Setting up the examples

Checkout this repository. The example directory include several examples of lambda function. All of them are built to get event from S3 (and write their results there)

The example directory contain:

* `mless.mlab` script - this script will start the mless container
* `ddns.dynu.sh` - this is a template for a script that will register to [dynu](https://www.dynu.com) dynamic DNS service. If this is what you are using copy it to `ddns.sh` and edit it to put your credentials and host name.
  When `mless.mlab` will start it will execute this script to register the updated IP.
  If you are using another service you can put another script there.

#### First example

Now create a lambda function to hold the first proxy:

* As the environment choose python 2.7
* Set in the environment `MLESS_SERVER` to be `http://<your servername>:8000`
* Configure a trigger on creating an object in some S3 bucket
  (you should probably limit this only to a specific folder).
* Make sure to configure the lambda function with a role that allow it to read/write the above bucket.

Start the test server by running the script `mlessd.mlab` in the examples directory.

Now upload a file to the S3 bucket to the specified folder. You should notice that:

* The mlessd server will be invoked (twice)
* In addition to the file that you uploaded there should be another file with the sha1 of the original file

What happens:

* The file was uploaded to S3 which triggered the lambda function
* The mlessd proxy code invoked the mlessd with the details of the original file
* The proxy was called, and executed the function from the local code.
  This code computed the hash and wrote it to S3
* Writing the hash to the S3 triggered the process again, AWS Lambda called the proxy
  which called mlessd which executed the function.
* This time the code identify by the filename extension that it doesn't need to write the hash,
  so it breaks the loop

#### Modifying the example

Open your favorite IDE and edit `example/hash/lambda_function.py`.
For example change the hashed file extension to be `.hash` by changing the line:

```python
newkey = key+".sha1"
```

to:

```python
newkey = key+".hash"
```

Save the file your (local) editor and upload another file to S3 (no need to stop/start mlessd).
You should notice that the new code is being called:

* The files are created with the `.hash` extension
* The code fails to identify the loops so you get files with `.hash.hash`, `.hash.hash.hash`, etc extension
  (the "loop" will continue until the S3 file name length limit is reached)

#### Running with debugger

Since we are in a proof of concept stage, currently only [pydevd](http://www.pydev.org) is supported.
It was chosen because it is free and has integration with an IDE (it is available as a free add on to eclipse).
Technically it is quite challenging since the process under debugging open a connection to the debugger running on the desktop.

In order to use it:

* Prepare a pydevd capable IDE (for example eclipse with the pydev add-on).
* Start the pydev server and set a breakpoint in the hashing function
* Edit `examples/template.yaml`, in the section describing the hashFile function, remove the comment from the debbuger line (and save the file)

Upload another file to S3 and watch the debugger in action.

#### More examples

The examples directory contain more examples that are documented in those directories.
