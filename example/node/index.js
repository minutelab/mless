'use strict';

console.log('Loading function');

const aws = require('aws-sdk');
const crypto = require('crypto');

const s3 = new aws.S3({ apiVersion: '2006-03-01' });


exports.handler = (event, context, callback) => {
    // console.log('Received event:', JSON.stringify(event, null, 2));

    const bucket = event.Records[0].s3.bucket.name;
    const key = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, ' '));
    console.log("invoked: " + bucket + ":" + key)
    console.log("remaining time: " + context.getRemainingTimeInMillis() )

    if (key.endsWith('.sha1')) {
        callback(null,"nothing to do")
        return
    }

    var params = {
        Bucket: bucket,
        Key: key,
    };
    s3.getObject(params, (err, data) => {
        if (err) {
            console.log(err);
            const message = `Error getting object ${key} from bucket ${bucket}. Make sure they exist and your bucket is in the same region as this function.`;
            console.log(message);
            callback(message);
            return
        }
        console.log('CONTENT TYPE:', data.ContentType);
        try {
            var hash = crypto.createHash('sha1');
            hash.update(data.Body)
            var digest = hash.digest('hex')
            params = {
                Bucket: bucket,
                Key: key + '.sha1',
                Body: digest
            }
            s3.putObject(params, (err, data) => {
                if (err) {
                    console.log(err)
                    callback(err)
                } else {
                    console.log('wrote object VersionId: ' + data.VersionId)
                    callback(null, 'hashed file: ' + digest);
                }
            })
        } catch (err) {
            callback(err)
        }
    });
};
