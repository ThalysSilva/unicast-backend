import { Injectable } from '@nestjs/common';
import { JwtService, JwtSignOptions } from '@nestjs/jwt';
import * as bcrypt from 'bcrypt';

import { doNothing } from 'src/utils/functions/general';
import {
  ApplicationError,
  NotAuthorizedError,
} from 'src/common/applicationError';
import { UserWithPassword } from 'src/@entities/user';
import { JwtPayload } from 'src/@types/jwt';
import { ConfigService } from '@nestjs/config';
import { omit } from 'lodash';
import { UserRepository } from 'src/repositories/userRepository';
import { pbkdf2Sync } from 'crypto';
import { EncryptJWT } from 'jose';

@Injectable()
export class AuthenticationService {
  private refreshTokenConfig: JwtSignOptions;
  private jweSecret: Uint8Array<ArrayBuffer>;

  constructor(
    private jwtService: JwtService,
    private userRepository: UserRepository,
    private configService: ConfigService,
  ) {
    this.refreshTokenConfig = this.configService.get('refresh-jwt') ?? {};
    const jweSecretHex = this.configService.get('JWE_SECRET_KEY') ?? '';
    const jweSecretBuffer = Buffer.from(jweSecretHex, 'hex');
    this.jweSecret = new Uint8Array(jweSecretBuffer);
  }

  async validateUser(
    email: string,
    password: string,
  ): Promise<UserWithPassword | null> {
    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      return null;
    }

    const userWithPassword = await this.userRepository.findByIdWithPassword(
      user.id,
    );

    if (!userWithPassword) {
      throw new NotAuthorizedError({
        action: 'authenticationService.validateUser',
        message: 'Usuário não encontrado',
      });
    }

    const passwordValid = await bcrypt.compare(
      password,
      userWithPassword.password,
    );

    if (!passwordValid) return null;

    return { ...user, password };
  }

  private async generateJwe(password: string, salt: string) {
    const smtpKey = pbkdf2Sync(
      password,
      Buffer.from(salt, 'hex'),
      10000,
      32,
      'sha256',
    );
    const jwe = await new EncryptJWT({ smtpKey: smtpKey.toString('hex') })
      .setProtectedHeader({ alg: 'dir', enc: 'A256GCM' })
      .encrypt(this.jweSecret);

    return jwe;
  }

  async login(user: UserWithPassword) {
    const payload = {
      email: user.email,
      userId: user.id,
    };

    const accessTokens = await this.generateAccessTokens(payload);
    const jwe = await this.generateJwe(user.password, user.salt);

    return {
      data: omit(user, [
        'refreshToken',
        'createdAt',
        'updatedAt',
        'salt',
        'password',
      ]),
      jwe,
      ...accessTokens,
    };
  }

  async logout(userId: string) {
    await this.userRepository.update(userId, { refreshToken: null });

    return;
  }

  async refreshToken(refreshToken: string) {
    const payload = this.verifyRefreshToken(refreshToken);
    const user = await this.userRepository.findById(payload.userId);
    if (!user) {
      throw new NotAuthorizedError({
        action: 'authenticationService.RenewToken',
        message: 'Usuário Não encontrado',
      });
    }

    if (refreshToken !== user.refreshToken) {
      throw new NotAuthorizedError({
        action: 'authenticationService.RenewToken',
        message: 'Token de renovação inválido ou expirado',
      });
    }

    const tokens = await this.generateAccessTokens(payload);

    return {
      data: omit(user, ['refreshToken', 'createdAt', 'updatedAt']),
      ...tokens,
    };
  }

  private async generateAccessTokens({ exp, iat, ...payload }: JwtPayload) {
    doNothing([exp, iat]);
    const token = this.jwtService.sign(payload);
    const refreshToken = this.jwtService.sign(payload, this.refreshTokenConfig);

    await this.userRepository.update(payload.userId, { refreshToken });

    return { token, refreshToken };
  }

  private verifyRefreshToken(token: string) {
    try {
      const payload = this.jwtService.verify<JwtPayload>(
        token,
        this.refreshTokenConfig,
      );
      return payload;
    } catch (erro) {
      if (erro instanceof ApplicationError) throw erro;
      throw new NotAuthorizedError({
        action: 'authenticationService.verifyRefreshToken',
        message: 'Token expirado ou inválido',
      });
    }
  }
}
