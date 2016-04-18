# WeChat Server

This is a basic implementation for a HTTP server used for the official accout on wechat

## Setup
* Make sure you've registered an official account with wechat on https://mp.weixin.qq.com/
  If you don't want to register one just for the testing, you can also create a sandbox account at http://mp.weixin.qq.com/debug/cgi-bin/sandboxinfo?action=showinfo&t=sandbox/index

* Configure the server URL and token, which will be used by wechat to verify your server.

* Write down the appID, appsecret of the official account.

* If you want to test the web based login mechanics, make sure the redirect URL is configured.

## Running the Test Server
This repo can be run as a testing server directly, make sure you follow the steps below

* Expose the appID, appsecret and token in environment variables: WECHAT_APP_ID, WECHAT_APP_SECRET, WECHAT_APP_TOKEN.

* Run the server with go run main.go, the server will listen on port 8080.

* Since wechat requires the server to be reachable on the public Internet, you can use tools such as ngrok to create a tunnel to your local server.

* Make sure the public server URL and domain is properly configured with the official account on wechat.

* Then you should be able to follow the official account and interact with it.

## User OpenID
When interacting with the official account, each user is assigned with an unique and stable ID called Open ID.

Note that the open ID will be different when the same user interacts with another official account so it cannot be used to identify the player accross multiple official accounts even if they're run by the same organization.

## Getting User Info
There're mainly 2 ways to get the public info about the user:

### Access Token
Provided the official account has enough access on wechat, it can request an access token with which the user info can be obtained.

To get the access token, make a request to 
```
GET https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid={WECHAT_APP_ID}&secret={WECHAT_APP_SECRET}
```

If the call succeeds, the response body will be a json object in the following format:

```json
{
	"access_token" : "string, the token itself",
	"expires_in"   : "int, expire time in seconds"
}
```

Note that the token will expire in roughly 2 hours and it needs to be refreshed before can be used again.

[Reference](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140183&token=)

With the access token, user info can be obtained by making a request to 
```
GET https://api.weixin.qq.com/cgi-bin/user/info?access_token={ACCESS_TOKEN}&openid={OPEN_ID}
```

For this to work, the user needs to follow the offocial accout first. Otherwise there's no way to get his open ID.

If the call succeeds, it returns a json object in the following format:

```json
{
	"subscribe"      : "int, 1 for followed, 0 otherwise",
	"openid" 	     : "string, user's openid",
	"nickname" 	     : "string, user's wechat nickname",
	"sex" 		     : "int, 1 for male, 0 for female",
	"city" 		     : "string, name of the user's city",
	"country" 	     : "string, name of the user's country",
	"province" 	     : "string, name of the user's province",
	"language" 	     : "string, name of the user's language",
	"headimgurl"     : "string, url of the user's image on wechat",
	"subscribe_time" : "int, when the user followed this official account",
	"unionid"	     : "string, an id used to identify the same player accross multiple official accounts",
	"remark" 	     : "",
	"groupid"	     : "",
}
```

[Reference](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140839&token=)

### Web Based Login
Another way to get the user's public info without requiring him to follow first, is to use web based login：

* The user is asked to open a link in the wechat client: 
```
https://open.weixin.qq.com/connect/oauth2/authorize?appid={WECHAT_APP_ID}&redirect_uri={REDIRECT_URL}&response_type=code&scope=snsapi_userinfo&state={STATE}#wechat_redirect
```

* The user is presented with a page asking if he wants to share the public info with the official account

* If the user agrees, the wechat client is redirected to make a GET request to REDIRECT_URL/?code={CODE}&state={STATE}

* If the user disagrees, the wechat client is redirected to make a GET request to REDIRECT_URL/?state={STATE} - the code param is removed

* Since REDIRECT_URL should be an endpoint in our own server, the server can use the code param to get a web access token from wechat:
```
GET https://api.weixin.qq.com/sns/oauth2/access_token?appid={WECHAT_APP_ID}&secret={WECHAT_APP_SECRET}&code={CODE}&grant_type=authorization_code
```

* If the call succeeds, the response is a json object of the following format:
```json
{ 
	"access_token" 	: "string, the token itself",
 	"expires_in"	: "int, expire time in seconds",
 	"refresh_token"	: "string, the refresh token",
 	"openid"		: "string, open ID of the calling user",
 	"scope"			: "string, scope of the current token"
 } 
``` 

* With the web access token, the user info can be obtained:
```
GET https://api.weixin.qq.com/sns/userinfo?access_token={WEB_ACCESS_TOKEN}&openid={OPEN_ID}
```

[Reference](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140842&token=)
