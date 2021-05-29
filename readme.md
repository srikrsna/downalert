# Trigger Down Alert - Call / SMS / EMail
```
Host:
go get github.com/pavansh/downalert
Create emp.json
That's it.
```

# Setup
```
Third Party: Twilio and MailGun
Required Env Variables
-> Token      
-> TwilioSID  
-> TwilioAuth 
-> TwilioPhone
-> MailDomain 
-> MailKey    
-> SenderEmail
```

# Usage
```
URL: https://<url>/notify
Method: POST
Body: {
    "url": "xyz.com",
    "token": "<securekey>",
    "source": "<your source>",
    "message": "<message for sms>",
    "emailbody": "<message for email>",
    "group": "<project name>",
    "status": "<Up or Down>",
    "mode": "Call",
}

Default Mode: SMS/EMAIL, specify "call" for call alert
```