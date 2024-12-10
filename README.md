STEPS
=====
1. A client sends a request to the POST /v1/tokens/password-reset endpoint containing the email address of the user whose password they want to reset.
2. If a user with that email address exists in the users table, and the user has already confirmed their email address by activating, then generate a cryptographically-secure high-entropy random token.
3. Store a hash of this token in the tokens table, alongside the user ID and a short (30-60 minute) expiry time for the token.
4. Send the original (unhashed) token to the user in an email. (You can craft the email however you wish)
5. If the owner of the email address didn’t request a password reset token, they can ignore the email. Otherwise, they can submit the token to the PUT /v1/users/password endpoint along with their new password. 
6. If the hash of the token exists in the tokens table and hasn’t expired, then generate a bcrypt hash of the new password and update the user’s record.
7. Delete all existing password reset tokens for the user.

SAMPLE:
=======
1. $ curl -X POST -d '{"email": "alice@example.com"}' localhost:4000/v1/tokens/password-reset
    {
        "message": "an email will be sent to you containing password reset instructions"
    }
2. $ BODY='{"password": "your new password", "token": "Y7QCRZ7FWOWYLXLAOC2VYOLIPY"}'
$ curl -X PUT -d "$BODY" localhost:4000/v1/users/password
    {
        "message": "your password was successfully reset"
    }
       
3. Test out your CORS implementation by writing a small JS program that using the fetch() JS API.
4. Create a video that demonstrates your API (especially the password reset feature) and the CORS testing feature.

Commands
========
@data.json=
{
    "username": "cahlil",
    "email": "2021154337@ub.edu.bz", 
    "password": "password"
}

curl --header "Content-Type: application/json" --request POST --data @data.json localhost:4000/api/v1/users

@data.json=
{
    "token": ""
}

curl -H "Content-Type: application/json" -X PUT -d @data.json localhost:4000/api/v1/users/activated

@data.json=
{
    "email": "2021154337@ub.edu.bz", 
}

curl -X POST -d @data.json localhost:4000/api/v1/tokens/password-reset

@data.json=
{
    "password": "password2",
    "token": ""
}

curl -X PUT -d @data.json localhost:4000/api/v1/users/password

Cross-Origin Resource Sharing (CORS)
====================================
If two URLs have the same protocol, domain and port, they share the same origin.
i.e.:
   URL A: http://foo.com/a          vs URL B: https://foo.com/a
   Cross Origin (http protocol vs https protocol)

   URL A: http://foo.com/a          vs URL B: http://www.foo.com/a
   Cross Origin (foo.com domain vs www.foo.com domain)

   URL A: http://foo.com/a          vs URL B: http://foo.com:4000/a
   Cross Origin (port 80 vs port 4000)

   URL A: http://www.foo.com/a      vs URL B: http://www.foo.com/b
   Same Origin

   URL A: http://www.foo.com/a      vs URL B: http://www.foo.com/a?id=7
   Same Origin

Web browsers implement a security mechanism that blocks cross-origin requests by default.
The same-origin policy prevents websites from different origins from accessing each others resources.
Our API will decide which origins are allowed to read or access our responses or else the client's browser will block them by default.
When a response comes in, the browser checks to see if the client is allowed to read the response or not.

Before we enable CORS, if we try to access any resources from a different port, the same-origin policy prevents the response from being read.
After we enable it, the browser will inspect the response header and allow the response to be read.

If the port for the basic server is :9000 or :9001, the browser will inspect the response header and notice that its origin is a trusted origin.
However, if we change it to :9002, the browser will prevent our APIs resources from being read or accessed.

Simple and Preflight CORS Requests
==================================
If the cross-origin request's:
- HTTP method is HEAD, GET, or POST __AND__
- headers are Accept, Accept-Language, Content-Language, Content-Type or forbidden headers __AND__
- Content-Type header is application/x-www-form-urlencoded, multipart/form-data, or text/plain,
then it is a simple cross-origin request.

If the request is not simple, the browser will send a preflight request before the actual request we want to send to check if our actual request is safe to send by asking the API server if it accepts certain headers.

A preflight request will always have an:
- HTTP OPTIONS method
- Origin header (where the request is coming from)
- Access-Control-Request-Method header (HTTP method the real request will use)

We use the fetch() JS API to access our authentication endpoint so we need to provide our credentials. This will trigger a preflight request since HTTP does not know about our Authentication header.