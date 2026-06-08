import { Component, type ErrorInfo, type ReactNode } from "react";
import { ErrorState } from "./ErrorState";

interface Props {
  children?: ReactNode;
  appName: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class MicrofrontendErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error(`Microfrontend Error [${this.props.appName}]:`, error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: "2rem", height: "100%", width: "100%" }}>
          <ErrorState 
            title={`App Failed: ${this.props.appName}`}
            description={this.state.error?.message || "An unexpected error occurred while loading this module."}
          />
        </div>
      );
    }

    return this.props.children;
  }
}
