import { Injectable } from '@nestjs/common';
import { JwtService, JwtSignOptions } from '@nestjs/jwt';
import * as bcrypt from 'bcrypt';

import { doNothing } from 'src/utils/functions/general';
import {
  ApplicationError,
  NotAuthorizedError,
} from 'src/common/applicationError';
import { User } from 'src/@entities/user';
import { JwtPayload } from 'src/@types/jwt';
import { ConfigService } from '@nestjs/config';
import { omit } from 'lodash';
import { UserRepository } from 'src/repositories/userRepository';

@Injectable()
export class AuthenticationService {
  private refreshTokenConfig: JwtSignOptions;

  constructor(
    private jwtService: JwtService,
    private userRepository: UserRepository,
    private configService: ConfigService,
  ) {
    this.refreshTokenConfig = this.configService.get('refresh-jwt') ?? {};
  }

  async validateUser(email: string, password: string): Promise<User | null> {
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

    return user;
  }

  async login(user: User) {
    const payload = {
      email: user.email,
      userId: user.id,
    };

    const tokens = await this.generateTokens(payload);

    return {
      data: omit(user, ['refreshToken', 'createdAt', 'updatedAt']),
      ...tokens,
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

    const tokens = await this.generateTokens(payload);

    return {
      data: omit(user, ['refreshToken', 'createdAt', 'updatedAt']),
      ...tokens,
    };
  }

  private async generateTokens({ exp, iat, ...payload }: JwtPayload) {
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
