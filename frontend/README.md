# BookRank Frontend

A modern React application for book ranking and recommendations built with Vite, Tailwind CSS, and React Router.

## Features

- **Authentication**: JWT-based login and registration
- **Book Management**: Add, view, edit, and delete books with metadata fetching
- **Book Comparison**: Interactive side-by-side book comparison interface
- **Recommendations**: Personalized book recommendations based on user preferences
- **Friends System**: Connect with other users and view their activity
- **Responsive Design**: Mobile-first design that works on all devices
- **Real-time Updates**: Optimistic UI updates for better user experience

## Tech Stack

- **React 18** with hooks and functional components
- **Vite** for fast development and building
- **React Router 6** for navigation and routing
- **Tailwind CSS** for responsive styling
- **React Hook Form** for form validation
- **Axios** for API communication
- **Heroicons** for consistent iconography

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── auth/           # Authentication components
│   ├── book/           # Book-related components
│   ├── common/         # Generic UI components
│   └── layout/         # Layout components
├── pages/              # Route components
│   ├── auth/           # Login/register pages
│   ├── books/          # Book management pages
│   ├── dashboard/      # Dashboard page
│   ├── friends/        # Friends page
│   └── recommendations/ # Recommendations page
├── context/            # React Context providers
├── hooks/              # Custom React hooks
├── services/           # API service functions
└── utils/              # Helper functions
```

## Getting Started

### Prerequisites

- Node.js 16+ and npm/yarn
- BookRank backend API running on port 8080

### Installation

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` to configure your API URL:
   ```
   VITE_API_URL=http://localhost:8080
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```

   The app will be available at `http://localhost:3000`

### Building for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

## Key Features

### Authentication

- Secure JWT-based authentication
- Form validation with proper error handling
- Automatic token refresh and logout
- Protected routes with auth guards

### Book Management

- Add books with automatic metadata fetching
- Search and filter book library
- Detailed book pages with ratings and statistics
- Edit and delete books (creator only)

### Book Comparison System

- Interactive side-by-side comparison interface
- Visual feedback for comparisons
- Comparison history tracking
- Intelligent book pairing for optimal ranking

### Recommendations Engine

- Personalized recommendations based on user behavior
- Genre-based recommendations
- Feedback system to improve recommendations
- Social recommendations from friends

### Friends System

- Add/remove friends by username
- View friend activity and rankings
- Social features coming soon (challenges, leaderboards)

## API Integration

The frontend integrates with the BookRank backend API:

- **Authentication**: Login, register, user profile
- **Books**: CRUD operations, search, metadata
- **Ratings**: Book comparisons and ELO rating system
- **Recommendations**: Personalized and genre-based suggestions
- **Friends**: Social connections and activity

All API calls include:
- JWT authentication
- Error handling with user-friendly messages
- Loading states and optimistic updates
- Request/response interceptors for token management

## State Management

Uses React Context API for:
- **AuthContext**: User authentication state and actions
- **BookContext**: Book library, filters, and comparison state
- **ToastContext**: Global notification system

## Styling

Built with Tailwind CSS featuring:
- **Responsive Design**: Mobile-first approach
- **Custom Components**: Reusable button, input, modal components
- **Color System**: Consistent primary/secondary color palette
- **Animations**: Smooth transitions and loading states
- **Accessibility**: ARIA labels and keyboard navigation

## Performance Optimizations

- **Code Splitting**: Lazy loading of route components
- **Optimistic UI**: Immediate updates with rollback on error
- **Request Optimization**: Parallel API calls where possible
- **Image Optimization**: Proper image loading and fallbacks

## Development

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

### Code Style

- Functional components with hooks
- Proper prop validation
- Consistent naming conventions
- Error boundaries for graceful error handling

## Contributing

1. Follow the existing code style and patterns
2. Add proper error handling and loading states
3. Ensure responsive design works on all devices
4. Add appropriate ARIA labels for accessibility
5. Test thoroughly before submitting

## Environment Variables

- `VITE_API_URL` - Backend API URL (default: http://localhost:8080)
- `NODE_ENV` - Environment (development/production)

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

Built with modern JavaScript features and Vite for optimal performance.