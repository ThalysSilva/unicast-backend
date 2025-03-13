export type SendMailData = {
  to: string;
  subject: string;
  body: string;
};

export type TransporterMail = {
  host: string;
  port: number;
  auth: {
    user: string;
    pass: string;
  };
};

export interface MailAdapter {
  createTransporter: (transporter: TransporterMail) => Promise<void>;
  sendMail: (data: SendMailData) => Promise<void>;
}
