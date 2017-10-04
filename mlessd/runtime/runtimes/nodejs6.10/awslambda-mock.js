var crypto = require('crypto')
var http = require('http');

var HANDLER =  process.env._HANDLER

var FN_NAME = process.env.AWS_LAMBDA_FUNCTION_NAME
var VERSION = process.env.AWS_LAMBDA_FUNCTION_VERSION
var MEM_SIZE = process.env.AWS_LAMBDA_FUNCTION_MEMORY_SIZE
var TIMEOUT = process.env.AWS_LAMBDA_FUNCTION_TIMEOUT
var REGION = process.env.AWS_REGION
var ACCOUNT_ID = process.env.AWS_ACCOUNT_ID
var ACCESS_KEY_ID = process.env.AWS_ACCESS_KEY_ID
var SECRET_ACCESS_KEY = process.env.AWS_SECRET_ACCESS_KEY
var SESSION_TOKEN = process.env.AWS_SESSION_TOKEN

function consoleLog(str) {
  process.stderr.write(formatConsole(str))
}

function systemLog(str) {
  process.stderr.write(formatSystem(str) + '\n')
}

function systemErr(str) {
  process.stderr.write(formatErr(str) + '\n')
}

// Don't think this can be done in the Docker image
process.umask(2)

var OPTIONS = {
  initInvokeId: uuid(),
  handler: HANDLER,
  suppressInit: true,
  credentials: {
    key: ACCESS_KEY_ID,
    secret: SECRET_ACCESS_KEY,
    session: SESSION_TOKEN,
  },
  contextObjects: { // XXX do we need this here
    // clientContext: '{}',
    // cognitoIdentityId: undefined,
    // cognitoPoolId: undefined,
  },
}

var invoked = false
var start = null

// InvokeServer implement an http server that listen to a single request at a time
// After recieving this single request it close the socket until it is armed again
class InvokeServer {
    // start the server, and listen to a single request
    start() {
        if (this.server) {
            throw Error("InvokeServer already started")
        }
        this.cb = null
        this.err = null
        this.msg = null

        this.server = http.createServer((request,response) => this._recieve(request,response))
        this.server.listen(8999)
    }

    _recieve(request,response) {
        let body = [];
        request.on('error', (err) => {
            systemLog('_recieve error:' + err)
            this._result(err)
        }).on('data', (chunk) => {
            body.push(chunk);
        }).on('end', () => {
            var err
            var msg
            try {
                body = Buffer.concat(body).toString();
                msg = JSON.parse(body)
            } catch (e) {
                response.writeHead(500,'Exception')
                err = e
            }
            response.end()
            this._result(err, msg)
        });
    }

    _result(err,msg) {
        this._close()
        if (this.cb != null) {
            this.cb(err, msg)
        } else {
            this.err = err
            this.msg = msg
        }
    }

    _close() {
        this.server.close((err) => {
            if (err) {
                systemLog('closed server error: ' + err)
            }
        })
        this.server = null
    }

    // wait the server to recieve the message, if the message arrived before wait is called,
    // the call back is called imedatly, otherwise we store the cb to be called when the message arive
    wait(cb) {
        if (this.msg != null) {
            setImmediate(cb, this.err, this.msg)
            this.msg = null
        } else {
            if (!this.server) {
                throw Error("not listening")
            }
            this.cb = cb
        }
    }
}


var invokeServer = new InvokeServer()
invokeServer.start()

module.exports = {
  initRuntime: function() {
        process.stdout.write(JSON.stringify({OK:true}))
        return OPTIONS
  },
  waitForInvoke: function(fn) {
    invokeServer.wait( (err,msg,response) => {
        if (err != null) {
            systemLog('recived message error: ' + err)
            process.exit(2)
        }
        invoked=true
        start = process.hrtime()
        deadline = new Date(msg.deadline)
        TIMEOUT = 100*(1+Math.floor(module.exports.getRemainingTime()/100))
        try {
            fn({
                invokeid: msg.context.aws_request_id,
                invokeId: msg.context.aws_request_id,
                credentials: {
                    key: ACCESS_KEY_ID,
                    secret: SECRET_ACCESS_KEY,
                    session: SESSION_TOKEN,
                },
                eventBody: JSON.stringify(msg.event),
                contextObjects: { // TODO
                    // clientContext: '{}',
                    // cognitoIdentityId: undefined,
                    // cognitoPoolId: undefined,
                },
                invokedFunctionArn: msg.context.invoked_function_arn,
            })
        } catch (err) {
            systemLog('User handler encountered an error: ' + err)
            setImmediate(module.exports.reportDone, msg.context.aws_request_id,"exception",JSON.stringify(err.toString()))
        }
    })
  },
  reportRunning: function(invokeId) {}, // eslint-disable-line no-unused-vars
  reportDone: function(invokeId, errType, resultStr) {
    if (!invoked) {
        return
    }
    var diffMs = hrTimeMs(process.hrtime(start))
    var billedMs = Math.min(100 * (Math.floor(diffMs / 100) + 1), TIMEOUT * 1000)
    systemLog('END RequestId: ' + invokeId)

    process.stdout.write(JSON.stringify({
        result:JSON.parse(resultStr),
        errortype:errType,
        invokeid:invokeId,
        errors: errType!=null,
        billing: {
            duration: diffMs,
            memory: MEM_SIZE,
            Used: Math.round(process.memoryUsage().rss / (1024 * 1024)),
        }
    }))
    invokeServer.start()
  },
  reportFault: function(invokeId, msg, errName, errStack) {
    systemErr(msg + (errName ? ': ' + errName : ''))
    if (errStack) systemErr(errStack)
  },
  reportUserInitStart: function() {},
  reportUserInitEnd: function() {},
  reportUserInvokeStart: function() {},
  reportUserInvokeEnd: function() {},
  reportException: function() {},
  getRemainingTime: function() {
    return deadline.getTime() - (new Date()).getTime()
    },
  sendConsoleLogs: consoleLog,
  maxLoggerErrorSize: 256 * 1024,
}

function formatConsole(str) {
  return str.replace(/^[0-9TZ:\.\-]+\t[0-9a-f\-]+\t/, '\033[34m$&\u001b[0m')
}

function formatSystem(str) {
  return '\033[32m' + str + '\033[0m'
}

function formatErr(str) {
  return '\033[31m' + str + '\033[0m'
}

function hrTimeMs(hrtime) {
  return (hrtime[0] * 1e9 + hrtime[1]) / 1e6
}

// Approximates the look of a v1 UUID
function uuid() {
  return crypto.randomBytes(4).toString('hex') + '-' +
    crypto.randomBytes(2).toString('hex') + '-' +
    crypto.randomBytes(2).toString('hex').replace(/^./, '1') + '-' +
    crypto.randomBytes(2).toString('hex') + '-' +
    crypto.randomBytes(6).toString('hex')
}
