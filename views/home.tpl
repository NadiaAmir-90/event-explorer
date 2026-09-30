<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <title>Event Explorer</title>

    <link rel="stylesheet" href="/static/css/style.css">
</head>

<body>

<!-- =========================
     NAVBAR
========================= -->

<header class="navbar">

    <div class="nav-container">

        <a href="/" class="brand">
            <span class="brand-logo">e.</span>
            <span class="brand-name">eventexplorer</span>
        </a>

        <nav>
            <a href="/" class="nav-link active">
                Discover
            </a>
        </nav>

    </div>

</header>


<!-- =========================
     NOTICE BAR
========================= -->

<div class="notice-bar">
    Live event data / Explore music and sports in your selected city.
</div>


<main>

    <!-- =========================
         HERO
    ========================= -->

    <section class="hero">

        <div class="hero-content">

            <div class="hero-text">

                <div class="eyebrow">
                    <span></span>
                    LESS SCROLLING. MORE GOING.
                </div>

                <h1>
                    A city of possibilities.
                    <br>
                    <strong>Find your next one.</strong>
                </h1>

                <p class="hero-description">
                    Discover music and sports in one place.
                    <br>
                    Choose your city. Find something worth heading out for.
                </p>

                <div class="hero-tags">
                    <span>Live music</span>
                    <span>Sports &amp; matchdays</span>
                    <span>One simple search</span>
                </div>

            </div>


            <!-- =========================
                 HERO CARD
            ========================= -->

            <div class="hero-card-wrapper">

                <div class="hero-card">

                    <div class="hero-card-top">
                        <span>THE CITY IS CALLING</span>
                        <span>01 / 02</span>
                    </div>

                    <div class="hero-card-title">
                        GO
                        <br>
                        <span>OUT.</span>
                    </div>

                    <div class="hero-card-bottom">

                        <span>
                            GOOD PLANS.
                            <br>
                            GREAT MEMORIES.
                        </span>

                        <span class="arrow">
                            ↗
                        </span>

                    </div>

                </div>

            </div>

        </div>


        <!-- =========================
             CITY SEARCH
        ========================= -->

        <section class="city-search-card">

            <div class="city-search-heading">

                <h2>
                    Where are we going?
                </h2>

                <p>
                    Start with a city, then explore what is on.
                </p>

            </div>


            <form
                class="city-form"
                action="/events"
                method="GET"
                id="event-search-form"
            >

                <div class="city-input-wrapper">

                    <label for="city">
                        Choose a city
                    </label>


                    <div class="city-input">

                        <span class="location-icon">
                            ◎
                        </span>

                        <input
                            id="city"
                            type="text"
                            placeholder="Search a city, e.g. Toronto"
                            autocomplete="off"
                        >

                    </div>


                    <!-- Google suggestions appear here -->

                    <div
                        id="location-suggestions"
                        class="location-suggestions"
                    ></div>


                    <!--
                        These are the actual values submitted
                        to /events.
                    -->

                    <input
                        type="hidden"
                        id="selected-city"
                        name="city"
                    >

                    <input
                        type="hidden"
                        id="selected-country"
                        name="countryCode"
                    >


                    <p class="input-help">
                        Type at least 3 characters and select a suggestion.
                    </p>

                </div>


                <div class="explore-wrapper">

                    <button
                        type="submit"
                        class="explore-button"
                        id="explore-button"
                    >
                        Explore events
                        <span>→</span>
                    </button>

                    <p class="explore-help">
                        Select a city from the suggestions.
                    </p>

                </div>

            </form>

        </section>

    </section>


    <!-- =========================
         JOURNEY SECTION
    ========================= -->

    <section class="journey section">

        <div class="section-eyebrow">
            A SMALL PROJECT. THE COMPLETE JOURNEY.
        </div>

        <h2>
            From a city to a ticket.
        </h2>

        <p class="section-description">
            Explore the expected experience before building it with Go and Beego.
        </p>


        <div class="steps">

            <article class="step">

                <div class="step-number">
                    01
                </div>

                <h3>
                    Pick a place
                </h3>

                <p>
                    Find a city with autocomplete.
                </p>

            </article>


            <article class="step">

                <div class="step-number">
                    02
                </div>

                <h3>
                    Find your event
                </h3>

                <p>
                    Browse music and sports together.
                </p>

            </article>


            <article class="step">

                <div class="step-number">
                    03
                </div>

                <h3>
                    View the details
                </h3>

                <p>
                    Follow the safe demo ticket flow.
                </p>

            </article>

        </div>


        <!-- =========================
             SAMPLE CITIES
        ========================= -->

        <div class="sample-cities">

            <div class="sample-text">

                <h3>
                    Ready-to-test sample cities
                </h3>

                <p>
                    Toronto, London and Dhaka can be used to test
                    the event listing page.
                </p>

            </div>


            <div class="sample-links">

                <a href="/events?city=Toronto&countryCode=CA">
                    Toronto
                    <span>↗</span>
                </a>

                <a href="/events?city=London&countryCode=GB">
                    London
                    <span>↗</span>
                </a>

                <a href="/events?city=Dhaka&countryCode=BD">
                    Dhaka
                    <span>↗</span>
                </a>

            </div>

        </div>

    </section>


    <!-- =========================
         API GUIDE
    ========================= -->

    <section class="api-section section">

        <div>

            <div class="section-eyebrow">
                FOR INTERNS
            </div>

            <h2>
                What connects the pages?
            </h2>

            <p>
                The API guide explains every route, provider,
                and expected response.
            </p>

        </div>


        <a href="#" class="api-button">
            Open API guide
            <span>→</span>
        </a>

    </section>

</main>


<!-- =========================
     FOOTER
========================= -->

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


<!-- =========================
     AUTOCOMPLETE JAVASCRIPT
========================= -->

<script src="/static/js/autocomplete.js"></script>

</body>
</html>