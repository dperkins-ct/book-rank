import React, { forwardRef } from 'react'
import clsx from 'clsx'

const Input = forwardRef(function Input({
  label,
  error,
  helperText,
  className = '',
  ...props
}, ref) {
  return (
    <div className="space-y-1">
      {label && (
        <label
          htmlFor={props.id || props.name}
          className="block text-sm font-medium text-gray-700"
        >
          {label}
        </label>
      )}
      <input
        ref={ref}
        className={clsx(
          'input',
          error && 'border-red-300 focus:border-red-500 focus:ring-red-500',
          className
        )}
        {...props}
      />
      {error && (
        <p className="text-sm text-red-600">{error}</p>
      )}
      {helperText && !error && (
        <p className="text-sm text-gray-500">{helperText}</p>
      )}
    </div>
  )
})

export default Input