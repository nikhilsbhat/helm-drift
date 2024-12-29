scan/code: ## scans code for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs . --scanners vuln

scan/binary: mock/publish ## scans binary for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy config --input dist/helm-drift_darwin_amd64_v1/helm-drift
