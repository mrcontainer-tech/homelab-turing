# Policy Reporter

Web UI and reporting tool for Kyverno policy reports.

## Chart Info

- **Chart**: policy-reporter/policy-reporter v3.7.1
- **App**: Policy Reporter v3.7.0
- **Source**: https://kyverno.github.io/policy-reporter/

## Access

- **URL**: https://policy-reporter.mrcontainer.nl
- **Port-forward**: `kubectl port-forward svc/policy-reporter-ui -n policy-reporter 8080:8080`

## Features

- Dashboard view of policy report results
- Filter by namespace, policy, and status
- View pass/fail/warn/error/skip counts
- Drill down into individual policy violations

## Related

- [Kyverno](../kyverno/) - Policy engine that generates the reports
