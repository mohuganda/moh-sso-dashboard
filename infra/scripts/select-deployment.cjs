// Trusted build selection shared by automatic release delivery and manual promotion.
const BUILD = '.github/workflows/build.yml';
const RELEASES = new Set(['.github/workflows/backend-release.yml', '.github/workflows/frontend-release.yml']);

async function selectDeployment({ github, context, env }) {
  const repo = context.repo;
  const repository = `${repo.owner}/${repo.repo}`;
  const trusted = run => run.conclusion === 'success' && run.event === 'push' &&
    run.head_repository?.full_name === repository;
  const isBuild = run => trusted(run) && run.head_branch === 'main' && run.path?.split('@')[0] === BUILD;
  async function tagCommit(tag) {
    let { data: ref } = await github.rest.git.getRef({ ...repo, ref: `tags/${tag}` });
    let object = ref.object;
    for (let depth = 0; object.type === 'tag' && depth < 5; depth++) {
      const { data } = await github.rest.git.getTag({ ...repo, tag_sha: object.sha });
      object = data.object;
    }
    if (object.type !== 'commit') throw new Error('Release tag must resolve to a commit');
    return object.sha;
  }
  async function pairedRelease(sha) {
    for (const [component, path] of [['backend', 'backend/VERSION'], ['frontend', 'frontend/apps/shell/package.json']]) {
      const { data } = await github.rest.repos.getContent({ ...repo, path, ref: sha });
      const content = Buffer.from(data.content, 'base64').toString();
      const version = component === 'backend' ? content.trim() : JSON.parse(content).version;
      if (!/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(version)) return false;
      const tag = `${component}/v${version}`;
      try {
        const { data: release } = await github.rest.repos.getReleaseByTag({ ...repo, tag });
        if (release.draft || release.prerelease || !release.published_at || await tagCommit(tag) !== sha) return false;
      } catch (error) {
        if (error.status === 404) return false;
        throw error;
      }
    }
    return true;
  }
  let run;
  let environment;
  const automatic = context.eventName !== 'workflow_dispatch' || env.RELEASE_ONLY === 'true';
  if (context.eventName === 'workflow_dispatch') {
    if (!/^[0-9]+$/.test(env.BUILD_RUN_ID)) throw new Error('Invalid build run ID');
    environment = env.TARGET_ENVIRONMENT;
    if (!['staging', 'production'].includes(environment)) throw new Error('Invalid environment');
    ({ data: run } = await github.rest.actions.getWorkflowRun({ ...repo, run_id: Number(env.BUILD_RUN_ID) }));
    if (!isBuild(run)) throw new Error('Only successful trusted main-branch image-build runs may deploy');
    if (env.RELEASE_ONLY === 'true' && !await pairedRelease(run.head_sha)) throw new Error('Both published releases must match the build commit');
  } else {
    let sha;
    const trigger = context.payload.workflow_run;
    if (context.eventName === 'workflow_run') {
      if (!trusted(trigger)) return null;
      const path = trigger.path?.split('@')[0];
      if (path !== BUILD && !RELEASES.has(path)) return null;
      sha = trigger.head_sha;
      if (isBuild(trigger)) run = trigger;
    } else if (context.eventName === 'release') {
      const release = context.payload.release;
      if (release.draft || release.prerelease || !/^(backend|frontend)\/v\d+\.\d+\.\d+$/.test(release.tag_name)) return null;
      sha = await tagCommit(release.tag_name);
    } else return null;
    if (env.AUTO_DEPLOY_PRODUCTION_ON_RELEASE === 'true' && await pairedRelease(sha)) {
      environment = 'production';
    } else if (env.AUTO_DEPLOY_STAGING === 'true' && run) {
      environment = 'staging';
    } else return null;
    if (!run) {
      const { data } = await github.rest.actions.listWorkflowRuns({ ...repo, workflow_id: 'build.yml',
        branch: 'main', event: 'push', status: 'success', head_sha: sha, per_page: 100 });
      run = data.workflow_runs.filter(candidate => isBuild(candidate) && candidate.head_sha === sha)
        .sort((a, b) => b.id - a.id)[0];
      if (!run) return null; // The successful main-build completion will check again.
    }
  }
  return { run_id: String(run.id), source_commit: run.head_sha, run_attempt: String(run.run_attempt),
    environment, automatic: String(automatic), should_deploy: 'true' };
}
module.exports = { selectDeployment };
