# hardblame
Use Hardenize.com to compile a list of domains with security excellence.

To see the results visit: [Hardblame](https://dotse.github.io/hardblame/)

## Data source
Hardblame uses data from the Hardenize dashboard [Sweden’s Hälsoläget](https://www.hardenize.com/dashboards/sweden-health-status/).

## How are points awarded?
Points are awarded in three categories DNS, EMAIL and WEB.

Every measurement can have four outcomes:

|Outcome|Points|
|---|---|
|good| 1 point|
|neutral| 0 points|
|warning|-1 point|
|error|-2 points|

## Measurements

|   |   |
|---|---|
|**DNS**|
|Nameservers|Check if name servers are working|
|DNSSEC|Zone data can be validated|
|   |   |
|**EMAIL**|
|TLS|Server supports TLS with reasonable security|
|DANE|TLSA records are published for the mail servers|
|SPF|A valid SPF record is published|
|DMARC|A valid DMarc record with a policy other than "none"|
|   |   |
|**WEB**|
|TLS|HTTPS support on web with reasonable security|
|CSP|Content Security Policy is found|
|Headers|Security Headers are found|
|Cookies|Cookies are marked as secure|
|Mixed|No mixed content|
|XSS|Cross Site Scripting protection is implemented|
