scan/code: ## scans code for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs /helm-drift

<<<<<<< HEAD
scan/binary: mock/publish ## scans binary for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy config --input dist/helm-drift_darwin_amd64_v1/helm-drift
=======
scan/binary: ## scans binary for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs /helm-drift/dist/helm-drift_darwin_amd64_v1/helm-drift --scanners vuln
>>>>>>> 860c293 (Update the trivy configurations)
