export type SendMailData = {
  to: string;
  subject: string;
  body: string;
};

export type TransporterMailData = {
  host: string;
  port: number;
  auth: {
    user: string;
    pass: string;
  };
};

export interface MailAdapter {
  setTransporterConfig: (data: TransporterMailData) => Promise<void>;
  sendMail: (data: SendMailData) => Promise<void>;
}
