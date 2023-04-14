## EWS Exchange Web Service Extend

Exchange Web Service client for golang, base on `mhewedy/ews` , if you feel useful,please give me a star, and if you find some bug, please send a issue or pull a request, thanks.

### Improvement:

- fixed some bug about `NTLM` (~~update `github.com/Azure/go-ntlmssp` package version~~ Change `github.com/Azure/go-ntlmssp` to `github.com/vadimi/go-http-ntlm/v2`,because `go-ntlmssp` has so many bugs....).
- fixed some service not support `HTTP1.1` ,but go will use `HTTP2.0` if our service URL is `HTTPS` (After `Go 1.6` ，See [THIS](https://pkg.go.dev/net/http)).
- Add support for email with Attachment.
- Using strategy like [ews-java-api](https://github.com/OfficeDev/ews-java-api) deal with email with attachment.which is save the mail content and subject,than save mail attachment,finally send the mail we saved.(Because _ews-java-api_ project comment said, see below)

```java
        // Bug E14:80316 -- If the message has attachments, save as a
        // draft (and add attachments) before sending.
```

### TODO:

- [ ] Inline Attachment - like picture in email body.
- [ ] Support html email body with inline attachment.
- [x] Support Mail CC and BCC.

### usage:

```go
package main

import (
	"fmt"
	"github.com/johnchenkzy/ews"
	"github.com/johnchenkzy/ews/ewsutil"
	"log"
)

func main() {
	conf := ews.GetDefaultConfig()

	c := ews.NewClient(
		"https://outlook.office365.com/EWS/Exchange.asmx",
		"email@exchangedomain",
		"password",
		&conf,
	)

	email := ewsutil.Email{
			To:          "johnchentest@gmail.com",
			Cc:          "someone@else.com",
			Bcc:         "someone@else.com",
			Attachments: []string {"./path/to/file1", "./path/to/file2"},
			Subject:     "An email subject",
			Body:        "The email body, as plain text",
			BodyType:    ewsutil.BodyTypeText,
		}
	return ewsutil.SendEmails(client, email)
}

```

> Note: if you are using an on-premises Exchange server (or even if you manage your servers at the cloud), you need to pass the username as `AD_DOMAINNAME\username` instead, for examle `MYCOMANY\mhewedy`.

### Supported Feature matrix:

| Category                         | Operation            | Supported\*                        |
| -------------------------------- | -------------------- | ---------------------------------- |
| eDiscovery operations            |                      |                                    |
| Exchange mailbox data operations |                      |                                    |
|                                  | CreateItem operation | ✔️ (Email & Calendar & Attachment) |
|                                  | GetUserPhoto         | ✔️                                 |
| CreateAttachment operation       | ✔️                   |                                    |
| Availability operations          |                      |                                    |
|                                  | GetUserAvailability  | ✔️                                 |
|                                  | GetRoomLists         | ✔️                                 |
| Bulk transfer operations         |                      |                                    |
| Delegate management operations   |                      |                                    |
| Inbox rules operations           |                      |                                    |
| Mail app management operations   |                      |                                    |
| Mail tips operation              |                      |                                    |
| Message tracking operations      |                      |                                    |
| Notification operations          |                      |                                    |
| Persona operations               |                      |                                    |
|                                  | FindPeople           | ✔️                                 |
|                                  | GetPersona           | ✔️                                 |
| Retention policy operation       |                      |                                    |
| Service configuration operation  |                      |                                    |
| Sharing operations               |                      |                                    |
| Synchronization operations       |                      |                                    |
| Time zone operation              |                      |                                    |
| Unified Messaging operations     |                      |                                    |
| Unified Contact Store operations |                      |                                    |
| User configuration operations    |                      |                                    |

- Not always 100% of fields are mapped.

### Extras

Besides the operations supported above, few new operations under the namespace `ewsutil` has been introduced:

- `ewsutil.SendEmail`
- `ewsutil.CreateEvent`
- `ewsutil.ListUsersEvents`
- `ewsutil.FindPeople`
- `ewsutil.GetUserPhoto`
- `ewsutil.GetUserPhotoBase64`
- `ewsutil.GetUserPhotoURL`
- `ewsutil.GetPersona`

NTLM is supported as well as Basic authentication

#### Reference:

https://docs.microsoft.com/en-us/exchange/client-developer/web-service-reference/ews-operations-in-exchange
