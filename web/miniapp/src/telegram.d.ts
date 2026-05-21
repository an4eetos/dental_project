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
  themeParams: Record<string, string>;
  colorScheme: "light" | "dark";
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp };
}
