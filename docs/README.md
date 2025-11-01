# Documentation System

This directory contains configuration for the automatic documentation generation system.

## Overview

The Agent SDK Go project uses GitHub Actions to automatically generate API documentation when a new release is created. The documentation is then pushed to a separate React website repository where it's displayed using a React-based documentation UI.

## How It Works

1. When a new release is created in GitHub, the `docs-release.yml` workflow is triggered
2. The workflow generates documentation using `pkgsite` (modern GoDoc)
3. The documentation is bundled with metadata and configuration
4. The workflow checks out the React website repository
5. The bundle is extracted to the website repo and committed
6. A repository dispatch event triggers the website to rebuild

## Setting Up the React Website

To set up your React website to receive and display the documentation:

1. Create a React application for your documentation website
2. Create a GitHub repository to host it (e.g., `pontus-devoteam/agent-sdk-website`)
3. Configure your React app to read documentation from the `public/docs/api/{version}` directory
4. Set up a GitHub Actions workflow in the website repo to react to the `update-documentation` event

## React Website Structure

The documentation will be pushed to your React website repository with the following structure:

```
public/
  docs/
    api/
      {version}/          # e.g., v1.0.0
        api-docs/         # Contains the HTML documentation
        godoc-config.json # Configuration for rendering
        package-metadata.json
        go-version.txt
        package-list.txt
      latest/             # Symlink to the latest version
      versions.json       # List of all available versions
```

## Required Secrets

You need to set up the following secret in your Agent SDK Go repository:

- `WEBSITE_REPO_TOKEN`: A GitHub Personal Access Token with write access to your website repository

## Configuration Files

- `godoc-config.json`: Contains information about the package structure and UI customization

## Example React Integration

Here's a simplified example of how your React app might consume this documentation:

```jsx
import React, { useState, useEffect } from 'react';

function DocumentationViewer({ version = 'latest' }) {
  const [versions, setVersions] = useState([]);
  const [config, setConfig] = useState(null);
  const [metadata, setMetadata] = useState(null);
  
  useEffect(() => {
    // Load versions list
    fetch('/docs/api/versions.json')
      .then(res => res.json())
      .then(data => setVersions(data.versions));
      
    // Load config and metadata for current version
    fetch(`/docs/api/${version}/godoc-config.json`)
      .then(res => res.json())
      .then(data => setConfig(data));
      
    fetch(`/docs/api/${version}/package-metadata.json`)
      .then(res => res.json())
      .then(data => setMetadata(data));
  }, [version]);
  
  // Example: Render package structure based on config
  if (!config || !metadata) return <div>Loading...</div>;
  
  return (
    <div>
      <h1>{config.title} Documentation</h1>
      <p>Version: {metadata.version}</p>
      <p>Generated: {metadata.generated_date}</p>
      
      <div className="version-selector">
        <select onChange={e => window.location.href = `/docs/${e.target.value}`}>
          {versions.map(v => (
            <option key={v.version} value={v.version}>{v.version} ({v.date})</option>
          ))}
        </select>
      </div>
      
      <div className="documentation-container">
        {/* Render documentation sections based on config */}
        {config.sections.map(section => (
          <div key={section.name} className="doc-section">
            <h2>{section.name}</h2>
            <ul>
              {section.packages.map(pkg => (
                <li key={pkg}>
                  <a href={`/docs/api/${version}/api-docs/github.com/Muhammadhamd/agent-sdk-go/${pkg}`}>
                    {pkg}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
}
```

## Customizing Documentation

You can customize how the documentation is displayed by modifying the `godoc-config.json` file in this directory. This file contains:

- Package organization into sections
- Example code references
- UI styling preferences

## Troubleshooting

If documentation fails to generate:

1. Check the GitHub Actions logs for errors
2. Ensure your repository has the `WEBSITE_REPO_TOKEN` secret set
3. Verify that the React website repository exists and is accessible
4. Check that your Go code is properly documented with comments 