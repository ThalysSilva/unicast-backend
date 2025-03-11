import { LoginDto } from './schemas/login';
import { AuthenticationController } from './authentication.controller';
import { Test, TestingModule } from '@nestjs/testing';
import { AuthenticationService } from './authentication.service';
import { mock } from 'jest-mock-extended';
import { CustomRequest } from 'src/@types/services/types';
import { JwtService } from '@nestjs/jwt';
import { UserRepository } from 'src/repositories/userRepository';
import { ConfigModule, ConfigService } from '@nestjs/config';
import refreshJwtConfig from './config/refresh-jwt.config';
import { User } from 'src/@entities/user';

const jwtServiceMock = mock<JwtService>();
const userRepositoryMock = mock<UserRepository>();
const configServiceMock = mock<ConfigService>();
const authenticationServiceMock = mock<AuthenticationService>();

describe('AuthenticationController', () => {
  let controller: AuthenticationController;

  const req = mock<CustomRequest>({
    user: {
      id: 'test',
      email: 'test@email.com',
      name: 'test',
      createdAt: new Date(),
      updatedAt: new Date(),
      refreshToken: 'test',
    },
  });

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        { provide: AuthenticationService, useValue: authenticationServiceMock },
        { provide: JwtService, useValue: jwtServiceMock },
        { provide: UserRepository, useValue: userRepositoryMock },
        {
          provide: ConfigService,
          useValue: configServiceMock,
        },
      ],
      imports: [
        ConfigModule.forRoot({
          load: [refreshJwtConfig],
        }),
      ],
      controllers: [AuthenticationController],
    }).compile();

    controller = module.get<AuthenticationController>(AuthenticationController);

    configServiceMock.get.mockImplementation((key: string) => {
      if (key === 'jwtConfig') {
        return {
          secret: 'secret',
          signOptions: {
            expiresIn: '1h',
          },
        };
      }
      if (key === 'refresh-jwt') {
        return {
          secret: 'secret',
          expiresIn: '7d',
        };
      }
    });
  });

  describe('login', () => {
    it('should call the authentication service', async () => {
      const loginDto: LoginDto = { email: 'test@email.com', password: 'test' };

      await controller.login(req, loginDto);

      expect(authenticationServiceMock.login).toHaveBeenCalledTimes(1);
      expect(authenticationServiceMock.login).toHaveBeenCalledWith(req.user);
    });

    it('should return a token when credentials are valid', async () => {
      const loginDto: LoginDto = { email: 'test@email.com', password: 'test' };
      const token = 'token';
      const refreshToken = 'refresh';

      authenticationServiceMock.login.mockResolvedValue({
        token,
        data: {} as User,
        refreshToken,
      });

      const response = await controller.login(req, loginDto);

      expect(response).toHaveProperty('token', token);
    });

    it('should throw an error when credentials are invalid', async () => {
      const loginDto: LoginDto = {
        email: 'invalid@email.com',
        password: 'invalid',
      };

      authenticationServiceMock.login.mockRejectedValue(
        new Error('Invalid credentials'),
      );

      await expect(controller.login(req, loginDto)).rejects.toThrow(
        'Invalid credentials',
      );
    });
  });

  describe('logout', () => {
    it('should call the authentication service', async () => {
      await controller.logout(req);

      expect(authenticationServiceMock.logout).toHaveBeenCalledTimes(1);
      expect(authenticationServiceMock.logout).toHaveBeenCalledWith(
        req.user?.id,
      );
    });

    it('should throw an error when user is not authenticated', async () => {
      const invalidReq = mock<CustomRequest>({ user: undefined });

      await expect(controller.logout(invalidReq)).rejects.toThrow(
        'Usuário não autenticado',
      );
    });
  });

  describe('refreshToken', () => {
    it('should call the authentication service', async () => {
      const refreshToken = 'test';

      await controller.refreshToken(refreshToken);

      expect(authenticationServiceMock.refreshToken).toHaveBeenCalledTimes(1);
      expect(authenticationServiceMock.refreshToken).toHaveBeenCalledWith(
        refreshToken,
      );
    });

    it('should return a new token when refresh token is valid', async () => {
      const refreshToken = 'newRefreshToken';
      const token = 'newToken';

      authenticationServiceMock.refreshToken.mockResolvedValue({
        data: {} as User,
        token,
        refreshToken,
      });

      const response = await controller.refreshToken(refreshToken);

      expect(response).toHaveProperty('token', token);
    });

    it('should throw an error when refresh token is invalid', async () => {
      const refreshToken = 'invalid';

      authenticationServiceMock.refreshToken.mockRejectedValue(
        new Error('Invalid refresh token'),
      );

      await expect(controller.refreshToken(refreshToken)).rejects.toThrowError(
        'Invalid refresh token',
      );
    });
  });
});
