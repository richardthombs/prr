import type { ExtensionAPI, ExtensionContext } from "@mariozechner/pi-coding-agent";

const PRIME_TIMEOUT_MS = 30_000;

async function runBdPrime(pi: ExtensionAPI, ctx: ExtensionContext, trigger: string) {
	try {
		const result = await pi.exec("bd", ["prime"], {
			timeout: PRIME_TIMEOUT_MS,
		});

		if (result.code !== 0 && ctx.hasUI) {
			const message = result.stderr.trim() || result.stdout.trim() || `bd prime exited with code ${result.code}`;
			ctx.ui.notify(`bd prime failed on ${trigger}: ${message}`, "error");
		}
	} catch (error) {
		if (ctx.hasUI) {
			const message = error instanceof Error ? error.message : String(error);
			ctx.ui.notify(`bd prime failed on ${trigger}: ${message}`, "error");
		}
	}
}

export default function (pi: ExtensionAPI) {
	pi.on("session_start", async (event, ctx) => {
		await runBdPrime(pi, ctx, `session_start/${event.reason}`);
	});

	pi.on("session_compact", async (_event, ctx) => {
		await runBdPrime(pi, ctx, "session_compact");
	});
}
