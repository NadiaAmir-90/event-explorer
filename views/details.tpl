<!DOCTYPE html>
<html lang="en">

<head>

    <meta charset="UTF-8">

    <meta
        name="viewport"
        content="width=device-width, initial-scale=1.0"
    >

    <title>
        {{if .Page.Event}}
            {{.Page.Event.Title}}
        {{else}}
            Event Details
        {{end}}
        - Event Explorer
    </title>

    <link rel="stylesheet" href="/static/css/style.css?v=4">

</head>

<body>

<header class="navbar">

    <div class="nav-container">

        <a href="/" class="brand">

            <span class="brand-logo">
                e.
            </span>

            <span class="brand-name">
                eventexplorer
            </span>

        </a>

        <nav>

            <a href="/" class="nav-link">
                Discover
            </a>

        </nav>

    </div>

</header>


<main class="details-page">

    <!-- BACK LINK -->

    <a
        href="/events"
        class="details-back"
    >
        ← Back to events
    </a>


    {{if .Page.Error}}

        <!-- ERROR -->

        <section class="details-error">

            <div class="section-eyebrow">
                EVENT ERROR
            </div>

            <h1>
                Event unavailable
            </h1>

            <p>
                {{.Page.Error}}
            </p>

            <a
                href="/"
                class="explore-button"
            >
                Find another event
                <span>→</span>
            </a>

        </section>


    {{else}}

        <!-- =========================
             DETAILS LAYOUT
        ========================== -->

        <section class="details-layout">


            <!-- =========================
                 LEFT SIDE
            ========================== -->

            <div class="details-main">


                <!-- EVENT IMAGE -->

                <div class="details-image-wrapper">

                    {{if .Page.Event.ImageURL}}

                        <img
                            src="{{.Page.Event.ImageURL}}"
                            alt="{{.Page.Event.Title}}"
                            class="details-image"
                        >

                    {{else}}

                        <div class="details-image-placeholder">
                            <span>EVENT</span>
                        </div>

                    {{end}}

                </div>


                <!-- EVENT TITLE -->

                <div class="details-heading">

                    <div class="details-category">
                        {{.Page.Event.Category}}
                    </div>

                    <h1>
                        {{.Page.Event.Title}}
                    </h1>

                </div>


                <!-- DESCRIPTION -->

                <section class="details-description">

                    <div class="section-eyebrow">
                        ABOUT THIS EVENT
                    </div>

                    <h2>
                        About this event
                    </h2>

                    {{if .Page.Event.Description}}

                        <p>
                            {{.Page.Event.Description}}
                        </p>

                    {{else}}

                        <p>
                            No description is available for this event.
                        </p>

                    {{end}}

                </section>


            </div>


            <!-- =========================
                 RIGHT SIDEBAR
            ========================== -->

            <aside class="details-sidebar">


                <div class="details-sidebar-eyebrow">
                    MAKE A PLAN
                </div>


                <h2>
                    The details
                </h2>


                <!-- WHEN -->

                <div class="details-info">

                    <div class="details-info-label">
                        WHEN
                    </div>

                    <div class="details-info-value">
                        {{.Page.Event.Date}}
                    </div>

                </div>


                <!-- WHERE -->

                <div class="details-info">

                    <div class="details-info-label">
                        WHERE
                    </div>

                    <div class="details-info-value">
                        {{.Page.Event.Location}}
                    </div>

                </div>


                <!-- CATEGORY -->

                <div class="details-info">

                    <div class="details-info-label">
                        CATEGORY
                    </div>

                    <div class="details-info-value">
                        {{.Page.Event.Category}}
                    </div>

                </div>


                <!-- TICKET -->

                <div class="details-ticket">

                    <a
                        href="/redirect/{{.Page.Event.ID}}"
                        class="ticket-button"
                    >

                        <span>
                            View tickets
                        </span>

                        <span>
                            ↗
                        </span>

                    </a>


                    <p>
                        Opens the official ticket page for this event.
                    </p>

                </div>


            </aside>


        </section>

    {{end}}

</main>


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