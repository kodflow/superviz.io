# Security Policy

## 🛡 Supported Versions

We actively maintain the latest **stable** version of Superviz.io. Older releases may receive critical patches on a case-by-case basis.

| Version       | Supported | Notes                 |
| ------------- | --------- | --------------------- |
| `main / HEAD` | ✅ Yes     | Actively maintained   |
| v1.x          | ✅ Yes     | Security patches only |
| < v1.0        | ❌ No      | Unsupported           |

## 📬 Reporting a Vulnerability

If you discover a security vulnerability in Superviz.io, please report it **privately** via email:

* **Email**: [TODO](mailto:TODO)
* **PGP**: \[coming soon]

Please include the following:

* Description of the vulnerability
* Steps to reproduce (if possible)
* Affected version(s) or commit hash
* Your contact details (optional)

**Do not** open a public issue for security concerns.

We aim to respond within **3 business days** and will coordinate disclosure if necessary.

## 🔐 Security Best Practices (when using Superviz.io)

To help ensure secure deployments of Superviz.io:

* Run in minimal, hardened containers (e.g., Alpine or distroless)
* Never expose Superviz.io directly to the internet
* Isolate configuration files and secrets using proper file permissions
* Use supervision groups and user namespaces where supported
* Always validate downloaded binaries (checksum or signature)

## 📦 Dependency Vulnerabilities

We periodically audit Go dependencies using:

* `govulncheck`
* `gosec`
* GitHub Security Advisories

Any critical issues will be patched as part of the next release cycle or in an emergency fix.

## 📄 Disclosure Policy

Responsible disclosure is strongly encouraged. We credit contributors who report issues unless anonymity is requested.

---

Thanks for helping make Superviz.io more secure!
