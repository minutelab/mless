# Based on github.com/lambci/docker-lambda/python2.7/run/runtime-mock.py version cf93693
from __future__ import print_function
import sys
import os
import random
import uuid
import time
import resource
import datetime
import json

# hijack stdout so it won't be taken by mistake
orig_stdout = os.fdopen(os.dup(1),"w")
os.dup2(2,1)

orig_stderr = sys.stderr

if os.environ.get('_MLESS_DEBUGGER') == "ptvsd":
    import ptvsd
    ptvsd.enable_attach(secret=None)


def eprint(*args, **kwargs):
    print(*args, file=orig_stderr, **kwargs)
    sys.stdout.flush()


def _random_invoke_id():
    return str(uuid.uuid4())


def jsonStreamReader(f):
    data = ""
    decoder = json.JSONDecoder()
    ex = None
    while True:
        next = f.readline()
        if next == "":
            break
        data = data + next
        while data != "":
            # print "decoding:", data
            idx = json.decoder.WHITESPACE.match(data, 0).end()
            data = data[idx:]
            if data == "":
                break
            try:
                obj, end = decoder.raw_decode(data)
                yield obj
                data = data[end:]
                ex = None
            except ValueError as e:
                if ex and ex.message == e.message:
                    raise ValueError("%s while decoding '%s'" % (e,data))
                ex = e
                break
    if ex:
        raise ValueError("%s while decoding '%s'" % (ex,data))


_GLOBAL_MEM_SIZE = os.environ.get('AWS_LAMBDA_FUNCTION_MEMORY_SIZE', '1536') # TODO
_GLOBAL_TIMEOUT = int(os.environ.get('AWS_LAMBDA_FUNCTION_TIMEOUT', '300')) # TODO

_GLOBAL_MODE = 'event' # Either 'http' or 'event'
_GLOBAL_SUPRESS_INIT = True # Forces calling _get_handlers_delayed()
_GLOBAL_DATA_SOCK = -1
_GLOBAL_INVOKED = False
_GLOBAL_ERRORED = False
_GLOBAL_START_TIME = None

_GLOBAL_MESSAGE_READER = jsonStreamReader(sys.stdin)

def report_user_init_start():
    return

def report_user_init_end():
    return

def report_user_invoke_start():
    return

def report_user_invoke_end():
    return

def receive_start():
    sys.stdout = orig_stderr
    sys.stderr = orig_stderr

    eprint("recieved start called")
    msg=next(_GLOBAL_MESSAGE_READER)
    eprint("start message:", msg)

    handler = msg["handler"]
    os.environ["_HANDLER"] = handler
    eprint("Setting handler:",msg["handler"])

    global _GLOBAL_CREDENTIALS
    _GLOBAL_CREDENTIALS = {
        "key": os.environ.get("AWS_ACCESS_KEY_ID"),
        "secret": os.environ.get("AWS_SECRET_ACCESS_KEY"),
        "session":os.environ.get("AWS_SESSION_TOKEN"),
    }

    for k in msg["env"]:
        eprint("setting env %s=%s" % (k,msg["env"][k]))
        os.environ[k] = msg["env"][k]

    eprint("Debugger: ", os.environ.get('_MLESS_DEBUGGER'))
    if os.environ.get('_MLESS_DEBUGGER') == "pydevd":
        dhost = os.environ.get('_MLESS_DEBUG_HOST')
        eprint("Need to start debugger: "+dhost)
        if os.environ.get('PATHS_FROM_ECLIPSE_TO_PYTHON') == None:
            os.environ['PATHS_FROM_ECLIPSE_TO_PYTHON'] = json.dumps( [ [os.environ.get('_MLESS_DESKTOP_SOURCES'),'/var/task']])
            eprint("Pathes: " + os.environ['PATHS_FROM_ECLIPSE_TO_PYTHON'])
        import pydevd
        pydevd.settrace(host=dhost, suspend=False, stdoutToServer=True, stderrToServer=True)
        eprint("after calling settrace")
    elif os.environ.get('_MLESS_DEBUGGER') == "ptvsd":
        eprint("Waiting for debugger to attach")
        ptvsd.wait_for_attach()
        eprint("Debugger attached")

    return (
        _random_invoke_id(),
        _GLOBAL_MODE,
        handler,
        _GLOBAL_SUPRESS_INIT,
        _GLOBAL_CREDENTIALS
    )

def report_running(invokeid):
    res = { "ok": True }
    print(json.dumps(res)+"\n",file=orig_stdout)
    orig_stdout.flush()
    return

def receive_invoke():
    eprint("receive_invoke:")

    msg=next(_GLOBAL_MESSAGE_READER)

    global _GLOBAL_INVOKED
    global _GLOBAL_START_TIME
    global _GLOBAL_DEADLINE

    _GLOBAL_INVOKED = True
    _GLOBAL_START_TIME = time.time()
    _GLOBAL_DEADLINE = msg.get("deadline")/1000

    context = msg['context']
    context_objs = {
        'clientcontext': context.get('client_context'),
    }

    identity=context.get('identity')
    if identity != None:
        context_objs['cognitoidentityid'] = identity['cognito_identity_id']
        context_objs['cognitopoolid'] = identity['cognito_identity_pool_id']

    eprint("Start RequestId: %s Version: %s" % (context['aws_request_id'], os.environ.get("AWS_LAMBDA_FUNCTION_VERSION") ))

    return (
        context['aws_request_id'],
        _GLOBAL_DATA_SOCK,
        _GLOBAL_CREDENTIALS,
        json.dumps(msg['event']),
        context_objs,
        context['invoked_function_arn'],
        None, # What do we for xray_trace_id?
    )

def report_fault(invokeid, msg, except_value, trace):
    global _GLOBAL_ERRORED

    _GLOBAL_ERRORED = True

    if msg and except_value:
        eprint('%s: %s' % (msg, except_value))
    if trace:
        eprint('%s' % trace)
    return

def report_done(invokeid, errortype, result):
    global _GLOBAL_INVOKED
    global _GLOBAL_ERRORED

    if _GLOBAL_INVOKED:
        eprint("END RequestId: %s" % invokeid)

        duration = int((time.time() - _GLOBAL_START_TIME) * 1000)
        billed_duration = min(100 * int((duration / 100) + 1), _GLOBAL_TIMEOUT * 1000)
        max_mem = int(resource.getrusage(resource.RUSAGE_SELF).ru_maxrss / 1024)

        res = {
            "result":    result,
            "errortype": errortype,
            "invokeid":  invokeid,         # TODO needed?
            "errors":    _GLOBAL_ERRORED,
            "billing": {
                "duration": duration,
                "memory": _GLOBAL_MEM_SIZE,
                "used": max_mem,
            }
        }
        print(json.dumps(res)+"\n",file=orig_stdout)
        orig_stdout.flush()
        _GLOBAL_ERRORED = False
    else:
        return

def report_xray_exception(xray_json):
    return

def log_bytes(msg, fileno):
    eprint(msg)
    return

def log_sb(msg):
    return

def get_remaining_time():
    return int(1000*(_GLOBAL_DEADLINE - time.time()))

def send_console_message(msg):
    eprint(msg)
    return
