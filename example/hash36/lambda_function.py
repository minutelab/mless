import urllib.parse
import boto3
import hashlib
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
    key = urllib.parse.unquote_plus(event['Records'][0]['s3']['object']['key'], encoding='utf-8')
    print("File: %s:%s" % (bucket, key))
    if key.endswith(".sha1"):
        print("Nothing to do")
        return

    try:
        response = s3.get_object(Bucket=bucket, Key=key)
        sha1 = hash_stream(response['Body'],hashlib.sha1())
        print("SHA1: " + sha1)
        newkey = key+".sha1"
        print(newkey)
        response = s3.put_object(Bucket=bucket,Key=newkey, Body=sha1)
        return sha1
    except Exception as e:
        print(e)
        print('Error getting object {} from bucket {}. Make sure they exist and your bucket is in the same region as this function.'.format(key, bucket))
        raise e
