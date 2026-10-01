# Event Explorer

A server-side rendered event discovery application built with Go and Beego. It allows users to search for events by city and country and view event details and ticket links.

## How to Run

### 1. Clone the GitHub Repository

```bash
git clone <https://github.com/NadiaAmir-90/event-explorer>
cd event-explorer
```

### 2. Set Environment Variables

The application requires API keys for Ticketmaster and Google Places.

```bash
export TICKETMASTER_API_KEY="your_ticketmaster_api_key"
export GOOGLE_PLACES_API_KEY="your_google_places_api_key"
```

### 3. Install Dependencies

Download the required Go dependencies:

```bash
go mod download
```

### 4. Run the Application

Start the Beego application:

```bash
bee run
```

The application will be available at:

```
http://localhost:8080
```

## Architecture

The application follows an MVC-style architecture with a service layer.

```
                    Browser
                       |
                       v
                    Router
                       |
                       v
                  Controllers
                       |
                       v
                   Services
                  /         \
                 /           \
                v             v
        Ticketmaster       Google Places
             API                API
                \
                 v
                Cache
                 |
                 v
               Models
                 |
                 v
               Views
                 |
                 v
              Browser
```

### Main Components

- **Controllers** – Handle HTTP requests, validate input, and prepare responses.
- **Services** – Contain application and business logic.
- **Models** – Define data structures, API clients, and caching functionality.
- **Views** – Provide server-side rendered HTML templates.
- **Routers** – Define application routes and connect them to controllers.
- **Static** – Contains CSS and JS files.


## Features

- Search for events by city and country.
- City autocomplete using the Google Places API.
- Fetch Music and Sports events from the Ticketmaster Discovery API.
- View detailed information about individual events.
- Redirect users to approved Ticketmaster ticket pages.
- Thread-safe in-memory caching.
- Server-side rendered HTML using Beego templates.

## Technologies Used

- Go 1.25
- Beego Framework
- Ticketmaster Discovery API
- Google Places API

## API Routes

### Event Listing

```
GET /events?city={city}&countryCode={countryCode}
```

Example:

```
GET /events?city=Toronto&countryCode=CA
```

### Event Details

```
GET /events/{eventId}
```

### Ticket Redirect

```
GET /events/{eventId}/ticket
```

### Location Autocomplete

```
GET /api/locations/autocomplete?input={input}&sessionToken={sessionToken}
```

### Location Details

```
GET /api/locations/{placeId}?sessionToken={sessionToken}
```

## Caching

The application uses a thread-safe in-memory cache implemented with Go's `sync.RWMutex` and a map.

The cache works as follows:

```
Request
   |
   v
Cache Lookup
   |
   +---- MISS ----> Fetch API ----> Store Result
   |                                  |
   |                                  v
   |                                Return
   |
   +---- HIT -----> Return Cached Result
```

The cache does not use automatic time-based expiration. Cached results remain available until they are explicitly cleared.

## Environment Variables

| Variable | Description |
| --- | --- |
| `TICKETMASTER_API_KEY` | Ticketmaster Discovery API key |
| `GOOGLE_PLACES_API_KEY` | Google Places API key |
| `TICKETMASTER_BASE_URL` | Optional Ticketmaster API base URL |

The default Ticketmaster API base URL is:

```
https://app.ticketmaster.com/discovery/v2
```