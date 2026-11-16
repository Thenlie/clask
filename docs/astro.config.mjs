// @ts-check
import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'
import starlightLinksValidator from 'starlight-links-validator'

// Deployed to GitHub Pages at https://thenlie.github.io/clask/. Internal
// links in content must include the /clask base; the links validator fails
// the build on any that don't resolve.
export default defineConfig({
  site: 'https://thenlie.github.io',
  base: '/clask',
  integrations: [
    starlight({
      title: 'clask',
      description: 'Cross-repo task lists from TASKS.md files.',
      favicon: '/favicon.svg',
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/Thenlie/clask' }],
      editLink: { baseUrl: 'https://github.com/Thenlie/clask/edit/main/docs/' },
      customCss: ['./src/styles/custom.css'],
      plugins: [
        starlightLinksValidator({
          // Examples point at the local clask serve address.
          exclude: ['http://127.0.0.1:7373/**', 'http://127.0.0.1:7373', 'http://localhost:5173'],
        }),
      ],
      sidebar: [
        {
          label: 'Start',
          items: ['start/install', 'start/quick-start'],
        },
        {
          label: 'Task files',
          items: ['format/overview', 'format/file-block', 'format/task-block', 'format/writing-rules'],
        },
        {
          label: 'Registry',
          items: ['registry/config', 'registry/scopes', 'registry/refs', 'registry/discovery'],
        },
        {
          label: 'Guides',
          items: [
            'guides/monorepo',
            'guides/onboard-many-repos',
            'guides/git-hook',
            'guides/piping',
            'guides/editor',
          ],
        },
        {
          label: 'Web UI',
          items: ['web-ui/overview', 'web-ui/screenshots'],
        },
        {
          label: 'HTTP API',
          items: ['api/overview', 'api/tasks', 'api/projects', 'api/check', 'api/open', 'api/errors'],
        },
        {
          label: 'Reference',
          items: [
            'reference/json',
            { label: 'CLI', collapsed: true, items: [{ autogenerate: { directory: 'reference/cli' } }] },
          ],
        },
      ],
    }),
  ],
})
