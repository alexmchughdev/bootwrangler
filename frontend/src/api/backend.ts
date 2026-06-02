export interface HealthStatus {
  status: string;
  version: string;
}

interface AppService {
  Health(): Promise<HealthStatus>;
  Version(): Promise<string>;
}

declare global {
  interface Window {
    go?: {
      app?: {
        Service?: AppService;
      };
    };
  }
}

export async function getHealth(): Promise<HealthStatus> {
  const service = window.go?.app?.Service;
  if (!service) {
    return {
      status: "preview",
      version: "frontend-only",
    };
  }

  return service.Health();
}
