# BookRank Frontend

The BookRank frontend is a modern React application built with Vite, Tailwind CSS, and React Router that provides an intuitive interface for book ranking and recommendations. The application features JWT-based authentication, comprehensive book management with metadata fetching, interactive book comparison tools, personalized recommendations, and social features including a friends system, all designed with a mobile-first responsive approach.

## Technology Stack

The application utilizes React 18 with hooks and functional components for modern development patterns, Vite for fast development and optimized building, React Router 6 for client-side navigation, and Tailwind CSS for responsive styling. Form handling is managed through React Hook Form with validation, API communication uses Axios with proper error handling, and Heroicons provides consistent iconography throughout the interface.

## Running the Frontend

To run the frontend, ensure Node.js 16+ is installed and the BookRank backend API is running on port 8080. Install dependencies with `npm install`, copy the environment configuration with `cp .env.example .env`, and configure the `VITE_API_URL=http://localhost:8080` variable in the `.env` file. Start the development server using `npm run dev` to access the application at http://localhost:3000.

## Key Features

The application provides secure JWT-based authentication with form validation and automatic token management, comprehensive book management including adding books with metadata fetching and detailed book pages with ratings, an interactive side-by-side comparison system for building personalized rankings, and a recommendation engine that learns from user behavior. Social features include friend connections and activity viewing, while the entire interface maintains responsive design principles with real-time updates and optimistic UI patterns.

## State Management and API Integration

The frontend uses React Context API for global state management including authentication, book library management, and notifications. All API integration includes JWT authentication, comprehensive error handling with user-friendly messages, loading states with optimistic updates, and request interceptors for automatic token management. The system maintains performance through code splitting, lazy loading, parallel API calls where possible, and proper image optimization.

## Development and Deployment

For production deployment, use `npm run build` to generate optimized assets in the dist directory. The application supports modern browsers including Chrome 90+, Firefox 88+, Safari 14+, and Edge 90+. Development follows functional component patterns with hooks, proper error boundaries, consistent naming conventions, and accessibility standards including ARIA labels and keyboard navigation support.