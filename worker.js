const template = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{host}} {{vcs}} {{url}}">
</head>
</html>`;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const vcs = env.VCS_TYPE || "git";
    const vcsURL = new URL(env.VCS_URL || "https://github.com/reloadlife");

    if (url.pathname === "/ping") {
      return new Response("ok", { status: 200 });
    }

    if (request.method !== "GET") {
      return new Response("Method Not Allowed", { status: 405 });
    }

    const redirectURL = new URL(vcsURL.origin + vcsURL.pathname + url.pathname);

    if (url.searchParams.get("go-get") !== "1" || url.pathname.length < 2) {
      return Response.redirect(redirectURL.toString(), 307);
    }

    const host = url.host + url.pathname;
    const htmlContent = template
      .replace("{{host}}", host)
      .replace("{{vcs}}", vcs)
      .replace("{{url}}", redirectURL.toString());

    return new Response(htmlContent, {
      headers: {
        "Content-Type": "text/html",
        "Cache-Control": "no-store",
      },
    });
  },
};
