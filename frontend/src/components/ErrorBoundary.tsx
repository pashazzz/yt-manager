import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('UI error:', error, info)
  }

  handleReload = () => {
    this.setState({ error: null })
    window.location.reload()
  }

  render() {
    if (this.state.error) {
      return (
        <div className="error-boundary">
          <h2>Что-то пошло не так</h2>
          <p>{this.state.error.message}</p>
          <button className="btn-primary" onClick={this.handleReload}>
            Перезагрузить
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
