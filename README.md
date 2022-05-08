# Phishing Auto Clicker

This is a fake-human for simulating not-that-secure human in a pentest lab.

- downloading attachments from email.

if ending in `.zip` , it will be unzipped with password `infected`.

if ending in `.docm .doc .xls .xlsm`, it will be directly opened.

- clicking on links.

Always click on the first found link start with `http://` or `https://`.

The link must start with `http` and follow the format above.

But you should be aware that, before and after the link, there must be a space.

# Configuration

JSON doesn't support comments, remove all the comments if you copy them.

Example:

```jsonc
{
  "protocol": "imaps",   // valid values: imaps, imap ; WE DO NOT SUPPORT POP3
  "server": "my-email-server.com:995",   // server:port
  "user_email": "zhangsan@my-email-server.com",    // user email
  "password": "MY-PASSWORD-HERE",   // user password
  "save_to": "C:\\Downloads",    // attachments will be saved to this directory, use absolute path only, ALWAYS USE A NEW EMPTY FOLDER, IT WILL GET REMOVED AFTER PROGRAM EXIT.
  "enableTLS": true,   // enable TLS, if protocol ends with `s`, this option is forced to true
  "noTlsVerification": 1,      // 1=skip verfication, 2=normal
}
```

## License

 phishingAutoClicker
 Copyright (C) 2022  Patmeow Limited
 
 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.
 
 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.
 
 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.

