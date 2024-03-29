# Phishing Auto Clicker

This is a fake-human for simulating not-that-secure human in a pentest lab.

- downloading attachments from email. (Use startup arg: `-m 1` )

if ending in `.zip` , it will be unzipped and the first file will be opened.

We DO NOT support encrypted zip file due to limitation of Golang library.

if ending in `.docm .doc .xls .xlsm`, it will be directly opened.

Linux executables must be end in `.elf`.

- clicking on links. (Use startup arg: `-m 2` )

Always click on the first found link start with `http://` or `https://`.

The link must start with `http` and follow the format above.

But you should be aware that, before and after the link, there must be a space.

if link is a file with supported extension, it will be downloaded and open.

if link is ended with "submit", a credential will be POST-ed.

if link is just a URL like a webpage, it will be opened. More specifically, in Windows, use IE to open.

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

This config file must be encrypted using `xorer` in this repo.

## Software licensing

Custom.

## Open Source License

phishingAutoClicker
Copyright (C) 2022 Patmeow Limited

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

## Deployment in Exchange Server Side

- Enable IMAP4 in Exchange Server refs
  to https://docs.microsoft.com/en-us/exchange/clients/pop3-and-imap4/configure-imap4?view=exchserver-2019
- Set `loginType` to `PlainTextLogin` , refs to https://techgenix.com/exchange-2019configure-imap-settings/
  and  https://docs.microsoft.com/en-us/exchange/troubleshoot/client-connectivity/imap-clients-logon-fail-coexist-exchange
- Instruction: https://techgenix.com/exchange-2019configure-imap-settings/