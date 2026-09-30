<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <title>{{.Page.City}} Events - Event Explorer</title>

    <link rel="stylesheet" href="/static/css/style.css">
</head>

<body>

<header class="navbar">
    <div class="nav-container">

        <a href="/" class="brand">
            <span class="brand-logo">e.</span>
            <span class="brand-name">eventexplorer</span>
        </a>

        <nav>
            <a href="/" class="nav-link">Discover</a>
        </nav>

    </div>
</header>


<main class="listing-page">

    <!-- PAGE HEADER -->
    <section class="listing-header">

        <div class="listing-header-content">

            <div class="section-eyebrow">
                EVENTS IN
            </div>

            <h1>
                {{.Page.City}}
            </h1>

            <p>
                Discover music, sports and live events happening in
                {{.Page.City}}.
            </p>

        </div>

        <div class="listing-actions">

            <a href="/" class="change-city">
                ← Change city
            </a>

        </div>

    </section>


    <!-- ============================= -->
    <!-- MUSIC -->
    <!-- ============================= -->

    <section class="event-section">

        <div class="event-section-header">

            <div>
                <div class="section-eyebrow">
                    LIVE MUSIC
                </div>

                <h2>
                    Music
                </h2>
            </div>

            <span class="event-count">
                {{if .Page.Music}}{{len .Page.Music}}{{else}}0{{end}} events
            </span>

        </div>


        {{if .Page.Music}}

        <div class="event-grid">

            {{range .Page.Music}}

            <article class="event-card">

                <!-- EVENT IMAGE -->

                <div class="event-image-wrapper">

                    {{if .ImageURL}}

                    <img
                        src="{{.ImageURL}}"
                        alt="{{.Title}}"
                        class="event-image"
                    >

                    {{else}}

                    <div class="event-image-placeholder">
                        <span>EVENT</span>
                    </div>

                    {{end}}

                </div>


                <!-- EVENT CONTENT -->

                <div class="event-card-content">

                    <div class="event-date">
                        {{.Date}}
                    </div>

                    <h3 class="event-title">
                        {{.Title}}
                    </h3>

                    <p class="event-location">
                        {{.Location}}
                    </p>

                    <a
                        href="/events/{{.ID}}"
                        class="event-details-link"
                    >
                        View details
                        <span>→</span>
                    </a>

                </div>

            </article>

            {{end}}

        </div>

        {{else}}

        <div class="empty-events">

            <h3>
                No music events found
            </h3>

            <p>
                There are currently no music events available in
                {{.Page.City}}.
            </p>

        </div>

        {{end}}

    </section>


    <!-- ============================= -->
    <!-- SPORTS -->
    <!-- ============================= -->

    <section class="event-section">

        <div class="event-section-header">

            <div>
                <div class="section-eyebrow">
                    SPORTS &amp; MATCHDAYS
                </div>

                <h2>
                    Sports
                </h2>
            </div>

            <span class="event-count">
                {{if .Page.Sports}}{{len .Page.Sports}}{{else}}0{{end}} events
            </span>

        </div>


        {{if .Page.Sports}}

        <div class="event-grid">

            {{range .Page.Sports}}

            <article class="event-card">

                <!-- EVENT IMAGE -->

                <div class="event-image-wrapper">

                    {{if .ImageURL}}

                    <img
                        src="{{.ImageURL}}"
                        alt="{{.Title}}"
                        class="event-image"
                    >

                    {{else}}

                    <div class="event-image-placeholder">
                        <span>EVENT</span>
                    </div>

                    {{end}}

                </div>


                <!-- EVENT CONTENT -->

                <div class="event-card-content">

                    <div class="event-date">
                        {{.Date}}
                    </div>

                    <h3 class="event-title">
                        {{.Title}}
                    </h3>

                    <p class="event-location">
                        {{.Location}}
                    </p>

                    <a
                        href="/events/{{.ID}}"
                        class="event-details-link"
                    >
                        View details
                        <span>→</span>
                    </a>

                </div>

            </article>

            {{end}}

        </div>

        {{else}}

        <div class="empty-events">

            <h3>
                No sports events found
            </h3>

            <p>
                There are currently no sports events available in
                {{.Page.City}}.
            </p>

        </div>

        {{end}}

    </section>


    <!-- ============================= -->
    <!-- BOTTOM CTA -->
    <!-- ============================= -->

    <section class="listing-cta">

        <div>

            <div class="section-eyebrow">
                KEEP EXPLORING
            </div>

            <h2>
                Looking for somewhere else?
            </h2>

            <p>
                Choose another city and discover what's happening there.
            </p>

        </div>

        <a href="/" class="explore-button">
            Choose another city
            <span>→</span>
        </a>

    </section>

</main>


<!-- FOOTER -->

<footer class="footer">

    <div class="footer-container">

        <div>
            Event Explorer
            <span>/</span>
            Beego internship preview
        </div>

        <div>
            HTML + CSS + JavaScript
            <span>/</span>
            API &amp; route guide
        </div>

    </div>

</footer>

</body>
</html>