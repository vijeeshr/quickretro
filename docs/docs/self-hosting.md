# Self-Hosting

Although the [demo app](https://demo.quickretro.app) has all the features and can be used as-is, it runs on low resources. The data is auto-deleted within 2 hours. It is recommended to self-host the app for better performance.

## Update Allowed-Origins
<!-- [foo heading](./#heading)  -->
As defined in [Configurations](configurations#allowed-origins), update the config setting with your site origin.

## Secure Redis Instance
It is recommended to secure your Redis instance, preferably with ACL enabled. Check out the <code>redis</code> directory, and sample docker compose files <code>compose.yml</code>, <code>compose.reverseproxy.yml</code>, <code>compose.demohosting.yml</code> etc, which are part of this repository for more details.

## Sample Compose files
Check out the sample docker compose files <code>compose.yml</code>, <code>compose.reverseproxy.yml</code>, <code>compose.demohosting.yml</code> etc, which are part of this repository for more details.
