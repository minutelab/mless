Mless - run serverless functions locally in real-life context
=============================================================

Mless enables running AWS Lambda functions in a dev-accessible environment, using the full context (events and triggers) available in the Lambda framework. It is designed to greatly improve the development and testing process of serverless functions.

Using Mless you can execute Lambda functions in a lab environment optimized for development purposes, yet run the code in full context of the actual Lambda environment:

With mless the functions under test are triggered by actual AWS lambda events and triggers such as:

 * Calls from API gateway
 * Any AWS Lambda supported event sources such as
   * Changes in S3 buckets
   * Kinesis Streams events

This allows a full simulation of an application flow involving serverless functions, while allowing developers to have full access to the serverless code they work on.

The solution involves a small serverless proxy inside AWS lambda.
It enables redirection of events and triggers into an accessible lab environment where the code can be developed more easily and efficiently.
Once the Lambda function concludes it's processing, the flow of the application it is part of, continues as if the code runs in Lambda.
A single test flow may include multiple simulated Lambda functions.

Running the function will use the role configured for the lambda function,
and its configured environment. Yet using the [Minute Lab framework](http://minutelab.io) the lambda function itself will run in a preconfigured lab environment, that enables:

 * Using the latest code requiring no extra deployment steps of the new code
   * For scripted languages (python, node) you can simply save the code in the IDE which will trigger the auto-deployment of the code
 * Enabling debugging of the serverless code using your local IDE

 Though handling the code will look and feel as if it runs locally, it will actually run in AWS,
 which grants it access to resources in the private VPC.

## Proof Of Concept

The code is currently in a proof of concept stage.


The following limitations should be expected:
 * Limited environment support
    * python - both 2.7 and 3.6 are supported
    * nodejs6.10 partially supported
      * You can proxy to nodeJS lambda functions running locally
      * Currently we don't have proxy in nodejs, so the proxy must be in python
      * Remote debugging available both with the legacy debugger and inspector protocols
 * No authentication/security is currently available for data transfer between the Lambda proxy and the lab environment where the code actually runs.

On the other hand -
 * Expect frequent updates and added features
 * We are eager to get feedback

## Supported environment

| Environment | Runtime | Debug | Proxy |
| ----------- | ------- | ----- | ----- |
| python 2.7  | :white_check_mark: |  :white_check_mark: pydevd, ptvsd | :white_check_mark: |
| python 3.6  | :white_check_mark: |  :white_check_mark: pydevd, ptvsd | :white_check_mark: |
| nodejs 6.10 | :white_check_mark: | :white_check_mark:                | :x:                |
| nodejs 4.3  | :x:                | :x:                               | :x:                |
| java        | :x:                | :x:                               | :x:                |
| .net        | :x:                | :x:                               | :x:                |

## Technology

Mless combines several technologies

* a small mless proxy module inside the lambda function to allow execution of code in development in a controled or lab environment.
  This proxy can be used in two ways:
   * It can be the whole lambda function. A such it transfers all triggered events to the lab environment
   * It can be used as library inside the real lambda function. This enables transfering only part of the events to the lab environment
* The current Mless example utilizes the [Minute Lab](http://minutelab.io) technology to provide a lab environment optimized for development purposes, inculding the look & feel of a local lab environment for code that runs in the cloud as well as interactive troubleshooting, monitoring and automatic code deployment.


## MLess vs SAM Local

There are many similarities between mless and [SAM local](https://aws.amazon.com/blogs/aws/new-aws-sam-local-beta-build-and-test-serverless-applications-locally/).
Both projects aim to run lambda function code in a local environment.

In fact mless shares some of the code base with SAM local.

There are several differences though:

* Simulated VS actual triggers: SAM local runs and triggers the function locally and runs under the local user AWS role.
To trigger a function locally, SAM local enables you to simulate an event and similar to how it would run in Lambda. For example, SAM local enables you to simulate a trigger originating from a file change in S3.

However SAM local won't allow you to react to real events or view a chain of events. Suppose a Lambda function that is triggered by a change in a S3 bucket and writes something into a Kinesis stream, which will then trigger another Lambda function.

SAM local enables you to simulate each of the events independently.

With mless actual events and triggers are used to activate a Lambda function in test and the entire chain of event will roll out as it would in production.

* Reuse of containers: SAM local runs each function in its own container, it doesn't reuse the same container to run several functions.
Therefore it does not allow to test the side effects of container reuses (an inherent Lambda framework functionality). Both the positive effects (like preparing cache) and the negative ones are essential for a complete test scenario.

## Usage

### The demo Environment Overview


The environment includes the following components
* An S3 bucket that will trigger the Lambda code into action
* Mless proxy Lambda function: rediects triggers and events to a Minute Lab server, in which the Lambda code runs
* A minute Lab environment to run the Lambda code in. This is an mless container running inside a Minute Lab host, running in AWS.

### Prerequisites and setup

#### Setting up Minute Lab

To best demonstrate the value of mless in the devlopment process, the current mless setup relies on a [MinuteLab](http://minutelab.io) lab environment. You will have to register and install the client to activate your private lab environment.
The [Quick Start Guide](http://docs.minutelab.io/user-guide/quickstart/) is a good starting point. Follow this guide to learn how to setup a Minute Lab domain, a host and a file share between your desktop and your host.


**Note:** For the purpose of mless you will need to set up a self-hosted Minute Lab domain (all explained in the quickstart guide) to allow settings of security groups for inbound access from Lambda into your lab environment.

#### Setting Inbound Access

The mless Lambda proxy would have to access the mless container running inside a Minute Lab host. For this you will have to configure a security group allowing inbound access to a designated port (by default 8000).

You will have to setup the security group in the EC2 console. You can attach it to the host either in the EC2 console or the Minute Lab console.

#### Setting dynamic DNS

To access the container where the tested code runs, the proxy code would have to know its IP. The easiest way is by using a dynamic DNS service. You can use any service you like, but the mless code contains script templates to use [DYNU](https://www.dynu.com).


Register to such a service to obtain a DNS name (and credentials that allow you to register to it)

### Setting up the examples

Checkout this repository. The example directory includes several examples of lambda functions. All of them are built to get events from S3 (and write their results in there)

The example directory contains:

* `mless.mlab` script - this script will start the mless container in Minute Lab
* `ddns.dynu.sh` - this is a template for a script that will register to [dynu](https://www.dynu.com) dynamic DNS service. If this is what you are using copy it to `ddns.sh` and edit it to put your credentials and host name.
  As `mless.mlab` starts it executes this script to register the updated IP.

  If you are using another service you can put another script there.

#### First example

Create a Lambda function to hold the first proxy:

* Select either python 2.7 or python 3.6 as the environment
* Copy the `procy/python<ver>/python.py` as the function content
* Set in the environment `MLESS_SERVER` to be `http://<your servername>:8000`
* Make the function call the appropriate local function
  * The examples contain two function that does the same in different run time encironment
    * python2.7 - `hashFile`
    * python3.6 - `hash36`
  * By default the proxy call a function with the same name as its own name.
  * The default can be over ridden by configuring the environment variable `MLESS_FUNCNAME`.
* Configure a trigger on creating an object in a S3 bucket of your choice
  (it is advised to limit this to a specific folder only).
* Make sure to configure the Lambda function with a role that allows it to read/write from/to this bucket.

Start the test server by running the script `mlessd.mlab` (using the Minute Lab client) in the examples directory.

Now upload a file to the S3 bucket to the specified folder. You will notice that:

* The mlessd server will be invoked (twice)
* In addition to the file that you uploaded there should be another file containing the sha1 of the original file

What happened:

* The file was uploaded to S3, which triggered the Lambda function
* The mlessd proxy code invoked the mlessd with the details of the original file
* The proxy was called, and executed the function from inside the lab environment.
  This code computed the hash and wrote it back to S3
* Writing the hash to the S3 bucket triggered the process again. AWS Lambda called the proxy
  which called mlessd which executed the function.
* This time the code determined (by the filename extension) that it doesn't need to write the hash,

  and broke the loop.

#### Modifying the example

Open your favorite IDE and edit your serverless code. To do that open the file `example/hash/lambda_function.py`(stored locally on your desktop) and cahnge it.

For example change the hashed file extension to be `.hash`
This is done by changing the line:

```python
newkey = key+".sha1"
```

to:

```python
newkey = key+".hash"
```

Save the file you just edited (locally). It will be uploaded to the running Minute Lab container automatically.

Now upload another file to S3 (no need to stop/start mlessd).
You will notice that the NEW code is used:

* The files are created with the `.hash` extension
* The code fails to activate the "loop protection" (as the extention name was changed...), which results with `.hash.hash`, `.hash.hash.hash`, etc extensions (the "loop" will continue until the S3 file name length limit is reached)


#### Running with a debugger

The instructions below are for running with [pydevd](http://www.pydev.org) (using [eclipse](https://www.eclipse.org/) or [LiClipse](http://www.liclipse.com)) as IDE.
(It is also possible to debug using [Visual Studio Code](visual studio code) (`ptvsd`))

Using pydevd is technically non-trivial since the process under debugging opens a connection to the debugger running on the desktop. Minute Lab addresses the challenges quite efficiently.

In order to use it:

* Prepare a pydevd capable IDE (for example eclipse with the pydev add-on).
* Start the pydev server and set a breakpoint in the hashing function
* Edit `examples/template.yaml`, in the section describing the hashFile function, remove the comment from the debbuger line (and save the file)

Upload another file to S3 and watch the debugger in action.

#### More examples

The examples directory contains more examples that are documented in those directories.
