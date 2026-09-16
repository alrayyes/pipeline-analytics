<script lang="ts">
</script>

<svelte:head>
	<title>Privacy &amp; disclaimer · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="text-2xl font-semibold">Privacy &amp; disclaimer</h1>

	<div class="prose prose-sm dark:prose-invert mt-6 max-w-none">
		<h2>What this instance collects</h2>
		<p>
			This is self-hosted software -- whoever runs it operates their own
			instance, with their own database, and nothing here is sent to a
			pipeline-analytics vendor because there isn't one. A running instance
			collects and retains, for each repository someone registers:
		</p>
		<ul>
			<li>
				The repository's identifier and forge (GitHub or Forgejo), and the
				Forgejo instance URL where applicable.
			</li>
			<li>
				CI run, job, and step history ingested from that forge's Actions
				API -- names, statuses, timings, and conclusions, used to compute
				the health scores and trends this app displays.
			</li>
			<li>
				The access token entered when registering the repository, used to
				call that forge's API on the operator's behalf.
			</li>
		</ul>

		<h2>How tokens are stored</h2>
		<p>
			A repository's access token is encrypted at rest (AES-256, under a key
			the operator supplies via <code>--encryption-key</code>) and is never
			displayed again after registration -- only a masked form (its last
			four characters) is shown in the repository list. This dashboard's own
			browser storage can also remember a token locally, per forge, to offer
			it again when registering another repository on the same forge; that
			copy never leaves the browser it was entered in.
		</p>

		<h2>Nothing sent to a third party</h2>
		<p>
			Aside from calling the forge APIs (GitHub's or a Forgejo instance's)
			that a registered token already has access to, this app makes no
			outbound calls with a repository's data. The release history page is
			the one exception: it calls GitHub's public releases API directly from
			the browser to show this project's own changelog, unrelated to any
			tracked repository's data.
		</p>

		<h2>Disclaimer</h2>
		<p>
			pipeline-analytics is free software, licensed under the
			<a
				href="https://github.com/alrayyes/pipeline-analytics/blob/main/LICENSE"
				>GNU Affero General Public License v3.0</a
			>. As that license states, it is provided <strong>with no warranty</strong>,
			to the extent permitted by applicable law -- there is no guarantee of
			fitness for any particular purpose, and no vendor is responsible for
			how a self-hosted instance is run. The operator running an instance,
			not any third party, is responsible for the data it collects and how
			it's secured.
		</p>
	</div>
</main>
