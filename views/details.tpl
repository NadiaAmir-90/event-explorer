<!DOCTYPE html>
<html lang="en">

<head>

    <meta charset="UTF-8">

    <meta
        name="viewport"
        content="width=device-width, initial-scale=1.0"
    >

    <title>{{.Page.Title}}</title>

    <link
        rel="stylesheet"
        href="/static/css/style.css"
    >

</head>

<body>

<header class="navbar">

    <a href="/" class="brand">
        <span class="brand-mark">e.</span>
        <span>eventexplorer</span>
    </a>

    <nav>
        <a href="/" class="nav-link">
            Discover
        </a>
    </nav>

</header>

<main class="page-container">

    <a
        href="/events"
        class="back-link"
    >
        ← Back to events
    </a>

    {{if .Page.Error}}

        <div class="error-message">
            {{.Page.Error}}
        </div>

    {{else}}

        <article class="details-card">

            {{if .Page.Event.ImageURL}}

                <img
                    src="{{.Page.Event.ImageURL}}"
                    alt="{{.Page.Event.Title}}"
                    class="details-image"
                >

            {{else}}

                <div class="details-placeholder">
                    Event image unavailable
                </div>

            {{end}}

            <div class="details-content">

                <span class="event-category">
                    {{.Page.Event.Category}}
                </span>

                <h1>
                    {{.Page.Event.Title}}
                </h1>

                <div class="details-meta">

                    <p>
                        <strong>Date</strong>
                        {{.Page.Event.Date}}
                    </p>

                    <p>
                        <strong>Location</strong>
                        {{.Page.Event.Location}}
                    </p>

                </div>

                {{if .Page.Event.Description}}

                    <div class="description">
                        <h2>About this event</h2>

                        <p>
                            {{.Page.Event.Description}}
                        </p>
                    </div>

                {{else}}

                    <div class="empty-message">
                        No description is available.
                    </div>

                {{end}}

                <a
                    href="/redirect/{{.Page.Event.ID}}"
                    class="button primary"
                >
                    View Tickets
                </a>

            </div>

        </article>

    {{end}}

</main>

<footer class="footer">
    <p>Event Explorer</p>
</footer>

</body>
</html>