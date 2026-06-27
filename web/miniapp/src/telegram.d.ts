interface TelegramWebAppUser {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
}

interface TelegramWebApp {
  initData: string;
  initDataUnsafe: { user?: TelegramWebAppUser };
  platform?: string;
  version?: string;
  ready(): void;
  expand(): void;
  MainButton: {
    text: string;
    show(): void;
    hide(): void;
    onClick(cb: () => void): void;
    offClick(cb: () => void): void;
    showProgress(leaveActive?: boolean): void;
    hideProgress(): void;
    enable(): void;
    disable(): void;
  };
  showAlert(message: string, cb?: () => void): void;
  openLink(url: string, options?: { try_instant_view?: boolean }): void;
  openInvoice(
    url: string,
    callback?: (status: "paid" | "cancelled" | "failed" | "pending") => void,
  ): void;
  onEvent(eventType: "invoiceClosed", callback: (data: { status: string }) => void): void;
  offEvent(eventType: "invoiceClosed", callback: (data: { status: string }) => void): void;
  themeParams: Record<string, string>;
  colorScheme: "light" | "dark";
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp };
}
