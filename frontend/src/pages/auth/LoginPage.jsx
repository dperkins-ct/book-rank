import React, { useEffect } from 'react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useAuth } from '../../context/AuthContext'
import Button from '../../components/common/Button'
import Input from '../../components/common/Input'
import { BookOpenIcon } from '@heroicons/react/24/outline'

function LoginPage() {
  const { login, loading, error, isAuthenticated } = useAuth()
  const location = useLocation()
  const from = location.state?.from?.pathname || '/dashboard'

  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
    setError,
    watch
  } = useForm()


  useEffect(() => {
    if (error) {
      setError('root', { message: error })
    }
  }, [error, setError])

  if (isAuthenticated) {
    return <Navigate to={from} replace />
  }

  const onSubmit = async (data) => {
    try {
      await login(data)
    } catch (err) {
      // Error is handled by context
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="flex justify-center">
            <BookOpenIcon className="h-12 w-12 text-primary-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to BookRank
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Don&apos;t have an account?{' '}
            <Link
              to="/register"
              className="font-medium text-primary-600 hover:text-primary-500"
            >
              Sign up
            </Link>
          </p>
        </div>

        <form
          className="mt-8 space-y-6"
          onSubmit={handleSubmit(onSubmit)}
        >
          <div className="space-y-4">
            <Input
              label="Username"
              type="text"
              autoComplete="username"
              {...register('username', {
                required: 'Username is required',
                minLength: {
                  value: 3,
                  message: 'Username must be at least 3 characters'
                }
              })}
              error={errors.username?.message}
            />

            <Input
              label="Password"
              type="password"
              autoComplete="current-password"
              {...register('password', {
                required: 'Password is required'
              })}
              error={errors.password?.message}
            />
          </div>

          {errors.root && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="text-sm text-red-700">
                {errors.root.message}
              </div>
            </div>
          )}

          <button
            type="submit"
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50"
            disabled={loading}
          >
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}

export default LoginPage