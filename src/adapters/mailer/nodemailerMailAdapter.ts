import SMTPTransport from 'nodemailer/lib/smtp-transport';
import { MailAdapter, SendMailData, TransporterMailData } from './MailAdapter';
import nodemailer from 'nodemailer';
import { BadRequestError } from 'src/common/applicationError';

export class NodemailerMailAdapter implements MailAdapter {
  private transporter: nodemailer.Transporter<
    SMTPTransport.SentMessageInfo,
    SMTPTransport.Options
  >;
  private transporterUser: string;

  async createTransporter(data: TransporterMailData) {
    if (!this.transporter) {
      this.transporter = nodemailer.createTransport(data);
      this.transporterUser = data.auth.user;
    }

    try {
      await this.transporter.verify();
    } catch (error) {
      throw new BadRequestError({
        message: 'Invalid transporter',
        action: 'createTransporter.verify',
        details: { error },
      });
    }
  }

  async sendMail(data: SendMailData): Promise<void> {
    if (!this.transporter) {
      throw new BadRequestError({
        message: 'Transporter not created',
        action: 'sendMail',
      });
    }
    await this.transporter.sendMail({
      from: this.transporterUser,
      ...data,
    });
  }
}
