document.addEventListener("DOMContentLoaded", function () {

    const input = document.getElementById("city");

    const suggestionsBox =
        document.getElementById("location-suggestions");

    const selectedCity =
        document.getElementById("selected-city");

    const selectedCountry =
        document.getElementById("selected-country");

    const form =
        document.getElementById("event-search-form");

    const button =
        document.getElementById("explore-button");


    if (!input || !suggestionsBox || !form) {
        return;
    }


    /*
     * A session token is reused while the user
     * is searching and selecting one place.
     */
    let sessionToken = createSessionToken();


    let debounceTimer = null;


    // ------------------------------------------
    // Generate Google session token
    // ------------------------------------------

    function createSessionToken() {

        if (window.crypto && crypto.randomUUID) {
            return crypto.randomUUID();
        }

        return (
            Date.now().toString(36) +
            Math.random().toString(36).substring(2)
        );
    }


    // ------------------------------------------
    // Clear selected location
    // ------------------------------------------

    function clearSelectedLocation() {

        selectedCity.value = "";
        selectedCountry.value = "";

    }


    // ------------------------------------------
    // Search while typing
    // ------------------------------------------

    input.addEventListener("input", function () {

        const value = input.value.trim();


        /*
         * The user started typing again,
         * so the previous selection is no longer valid.
         */
        clearSelectedLocation();


        if (value.length < 3) {

            suggestionsBox.innerHTML = "";

            return;
        }


        clearTimeout(debounceTimer);


        debounceTimer = setTimeout(function () {

            fetchSuggestions(value);

        }, 300);

    });


    // ------------------------------------------
    // Fetch Google autocomplete suggestions
    // ------------------------------------------

    async function fetchSuggestions(inputValue) {

        try {

            const response = await fetch(
                "/api/locations/autocomplete?input=" +
                encodeURIComponent(inputValue) +
                "&sessionToken=" +
                encodeURIComponent(sessionToken)
            );


            if (!response.ok) {

                throw new Error(
                    "Unable to load city suggestions."
                );

            }


            const data = await response.json();

            renderSuggestions(data.suggestions || []);

        } catch (error) {

            console.error(error);

            suggestionsBox.innerHTML = `
                <div class="suggestion-error">
                    Unable to load suggestions.
                </div>
            `;

        }

    }


    // ------------------------------------------
    // Render suggestions
    // ------------------------------------------

    function renderSuggestions(suggestions) {

        suggestionsBox.innerHTML = "";


        if (!suggestions.length) {

            suggestionsBox.innerHTML = `
                <div class="suggestion-empty">
                    No cities found.
                </div>
            `;

            return;
        }


        suggestions.forEach(function (suggestion) {

            const item = document.createElement("button");

            item.type = "button";

            item.className = "location-suggestion";


            item.textContent =
                suggestion.text;


            item.addEventListener(
                "click",
                function () {

                    selectSuggestion(suggestion);

                }
            );


            suggestionsBox.appendChild(item);

        });

    }


    // ------------------------------------------
    // Select a Google suggestion
    // ------------------------------------------

    async function selectSuggestion(suggestion) {

        const placeId = suggestion.placeId;


        if (!placeId) {
            return;
        }


        input.value = suggestion.text;

        suggestionsBox.innerHTML = "";


        try {

            const response = await fetch(
                "/api/locations/" +
                encodeURIComponent(placeId) +
                "?sessionToken=" +
                encodeURIComponent(sessionToken)
            );


            if (!response.ok) {

                throw new Error(
                    "Unable to load selected city."
                );

            }


            const location = await response.json();


            /*
             * These values will be submitted
             * to /events.
             */
            selectedCity.value = location.city;
            selectedCountry.value = location.countryCode;


        } catch (error) {

            console.error(error);

            clearSelectedLocation();

            alert(
                "Unable to select this city. Please try again."
            );

        }

    }


    // ------------------------------------------
    // Submit form
    // ------------------------------------------

    form.addEventListener("submit", function (event) {

        const city = selectedCity.value.trim();

        const country = selectedCountry.value.trim();


        /*
         * User must select a Google suggestion.
         */
        if (!city || !country) {

            event.preventDefault();

            alert(
                "Please select a city from the suggestions."
            );

            return;
        }


        button.disabled = true;

        button.innerHTML =
            'Exploring... <span>→</span>';

    });

});