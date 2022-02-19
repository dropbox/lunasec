/*
 * Copyright by LunaSec (owned by Refinery Labs, Inc)
 *
 * Licensed under the Business Source License v1.1
 * (the "License"); you may not use this file except in compliance with the
 * License. You may obtain a copy of the License at
 *
 * https://github.com/lunasec-io/lunasec/blob/master/licenses/BSL-LunaTrace.txt
 *
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import React, { useState } from 'react';
import { Col, Container, Dropdown, Row } from 'react-bootstrap';
import { Filter } from 'react-feather';
import { useNavigate } from 'react-router-dom';

import { VulnerablePackageItem } from './VulnerablePackageItem';
import { Finding, severityOrder, VulnerablePackage } from './types';
interface FindingListProps {
  findings: Finding[];
}

function groupByPackage(findings: Finding[]): VulnerablePackage[] {
  // a place to group vulnerabilities by package
  const pkgs: Record<string, VulnerablePackage> = {};
  // sort by severity
  const sFindings = [...findings].sort((a, b) => severityOrder.indexOf(b.severity) - severityOrder.indexOf(a.severity));
  sFindings.forEach((f) => {
    const preExisting = pkgs[f.purl];
    if (!preExisting) {
      return (pkgs[f.purl] = {
        created_at: f.created_at, // might be better to sort and show the first date
        purl: f.purl,
        locations: f.locations,
        severity: f.severity,
        version: f.version,
        language: f.language,
        type: f.type,
        package_name: f.package_name,
        cvss_score: f.vulnerability.namespace === 'nvd' ? f.vulnerability.cvss_score : null,
        fix_state: f.fix_state || null,
        fix_versions: f.fix_versions || [],
        findings: [f],
      });
    }
    // just add the new vuln to the existing pkg, unless its a dupe from another location
    if (preExisting.findings.filter((pf) => pf.vulnerability.slug === f.vulnerability.slug).length === 0) {
      pkgs[f.purl].findings.push(f);
    }
    // add any new locations
    (f.locations as string[]).forEach((l) => {
      if (!preExisting.locations.includes(l)) {
        preExisting.locations = [...preExisting.locations, l];
      }
    });
    // set the highest CVSS score to the package score
    try {
      if (f.vulnerability.namespace === 'nvd' && Number(f.vulnerability.cvss_score) > Number(preExisting.cvss_score)) {
        preExisting.cvss_score = f.vulnerability.cvss_score;
      }
    } catch {
      console.error('failed converting cvss to number');
    }
    // Add any fix versions/state
    if (f.fix_state === 'fixed') {
      preExisting.fix_state = f.fix_state;
      const newFixedVersions = f.fix_versions as string[] | undefined;
      if (newFixedVersions) {
        newFixedVersions.forEach((v) => {
          if (!preExisting.fix_versions.includes(v)) {
            preExisting.fix_versions = [...preExisting.fix_versions, v];
          }
        });
      }
    }
  });
  return Object.values(pkgs);
  // console.log('sorted findings are ', sFindings)
}

export const VulnerablePackageList: React.FunctionComponent<FindingListProps> = ({ findings }) => {
  console.log('rendering finding list');
  const [severityFilter, setSeverityFilter] = useState(severityOrder.indexOf('Critical'));
  const prettySeverity = severityOrder[severityFilter] === 'Unknown' ? 'None' : severityOrder[severityFilter];

  const vulnerablePkgs = groupByPackage(findings);
  const filteredVulnerablePkgs = vulnerablePkgs.filter((pkg) => severityOrder.indexOf(pkg.severity) >= severityFilter);

  const pkgCards = filteredVulnerablePkgs.map((pkg) => {
    return (
      <Row key={pkg.purl}>
        <VulnerablePackageItem severityFilter={severityFilter} pkg={pkg} />
      </Row>
    );
  });

  return (
    <Container className="vulnerability-list">
      <Row>
        <Col md="6">
          <h2>Vulnerable Packages</h2>
        </Col>
        <Col md="6" style={{ display: 'flex', justifyContent: 'right' }}>
          <Dropdown align={{ md: 'end' }} className="d-inline me-2">
            <Dropdown.Toggle>Minimum Severity: {prettySeverity}</Dropdown.Toggle>
            <Dropdown.Menu>
              {severityOrder
                .map((severityName, severityIndex) => {
                  return (
                    <Dropdown.Item
                      active={severityIndex === severityFilter}
                      onClick={() => setSeverityFilter(severityIndex)}
                      key={severityIndex}
                    >
                      {severityName === 'Unknown' ? 'None' : severityName}
                    </Dropdown.Item>
                  );
                })
                .reverse()}
            </Dropdown.Menu>
          </Dropdown>
        </Col>
      </Row>
      <br />
      {pkgCards}
      {vulnerablePkgs.length > filteredVulnerablePkgs.length ? (
        <Row className="text-center">
          {' '}
          <span className="link-primary cursor-pointer" onClick={() => setSeverityFilter(0)}>
            Show {vulnerablePkgs.length - filteredVulnerablePkgs.length} lower severity vulnerabilities...{' '}
          </span>
        </Row>
      ) : null}
    </Container>
  );
};