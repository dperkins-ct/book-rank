import { useState, useEffect } from 'react'

export function useApi(apiFunction, dependencies = []) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const execute = async (...args) => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiFunction(...args)
      setData(response.data)
      return response.data
    } catch (err) {
      setError(err.response?.data?.message || err.message)
      throw err
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    execute()
  }, dependencies)

  return { data, loading, error, execute, refetch: execute }
}

export function useAsyncOperation() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const execute = async (asyncFunction) => {
    try {
      setLoading(true)
      setError(null)
      const result = await asyncFunction()
      return result
    } catch (err) {
      setError(err.response?.data?.message || err.message)
      throw err
    } finally {
      setLoading(false)
    }
  }

  return { loading, error, execute, setError }
}