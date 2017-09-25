from __future__ import print_function

import json
import os
import urllib
import boto3
import hashlib
import proxy
import re

s3 = boto3.client('s3')

BUF_SIZE=4096

def hash_stream(stream, hfunc):
    while True:
        data = stream.read(BUF_SIZE)
        if not data:
            break
        hfunc.update(data)
    return hfunc.hexdigest()

def hashFile(event, context):
    bucket = event['Records'][0]['s3']['bucket']['name']
    key = urllib.unquote_plus(event['Records'][0]['s3']['object']['key'].encode('utf8'))
    print("File: %s:%s" % (bucket, key))

    proxyPattern = os.environ.get("PROXY_PATTERN")
    if proxyPattern and re.match(proxyPattern,key):
        return proxy.proxy(event,context, funcName="hashProxy")

    if key.endswith(".md5") or key.endswith(".sha1"):
        print("Nothing to do")
        return

    try:
        response = s3.get_object(Bucket=bucket, Key=key)

        md5 = hash_stream(response['Body'],hashlib.md5())
        s3.put_object(Bucket=bucket,Key=key+".md5", Body=md5)

        sha1 = hash_stream(response['Body'],hashlib.sha1())
        s3.put_object(Bucket=bucket,Key=key+".sha1", Body=sha1)

        print("MD5: %s SHA1: %s" % (md5,sha1))

        return { "md5": md5, "sha1":sha1 }

    except Exception as e:
        print(e)
        print('Error getting object {} from bucket {}. Make sure they exist and your bucket is in the same region as this function.'.format(key, bucket))
        raise e
