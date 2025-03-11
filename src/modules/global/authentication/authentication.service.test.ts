import { Test, TestingModule } from '@nestjs/testing';
import { JwtService } from '@nestjs/jwt';
import { UserRepository } from 'src/repositories/userRepository';
import { AuthenticationService } from './authentication.service';
import { NotAuthorizedError } from 'src/common/applicationError';
import { User } from 'src/@entities/user';
import { JwtPayload } from 'src/@types/jwt';
import { mock } from 'jest-mock-extended';
import { ConfigModule, ConfigService } from '@nestjs/config';
import refreshJwtConfig from './config/refresh-jwt.config';

const jwtServiceMock = mock<JwtService>();
const userRepositoryMock = mock<UserRepository>();
const configServiceMock = mock<ConfigService>();

describe('AuthenticationService', () => {
  let service: AuthenticationService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthenticationService,
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
    }).compile();

    service = module.get<AuthenticationService>(AuthenticationService);
    configServiceMock.get.mockImplementation((key: string) => {
      if (key === 'jwtConfig') {
        return {
          secret: 'secret',
          signOptions: {
            expiresIn: '1h',
          },
        };
      }
      // Outras configurações
    });
  });

  describe('refreshToken', () => {
    it('should return new tokens for a valid refresh token', async () => {
      const refreshToken = 'validRefreshToken';
      const payload: JwtPayload = {
        userId: '123',
        email: 'user@email.com',
        exp: 1234567890,
        iat: 1234567890,
      };
      const user: User = {
        id: '123',
        refreshToken,
        email: '',
        name: '',
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      jwtServiceMock.verify.mockReturnValue(payload);
      userRepositoryMock.findById.mockResolvedValue(user);
      jwtServiceMock.sign.mockReturnValueOnce('newToken');
      jwtServiceMock.sign.mockReturnValueOnce('newRefreshToken');
      userRepositoryMock.update.mockResolvedValue(null);

      const result = await service.refreshToken(refreshToken);

      expect(result).toEqual({
        data: expect.any(Object),
        token: 'newToken',
        refreshToken: 'newRefreshToken',
      });
    });

    it('should throw NotAuthorizedError if user is not found', async () => {
      const refreshToken = 'validRefreshToken';
      const payload = { userId: '123' };

      jwtServiceMock.verify.mockReturnValue(payload);
      userRepositoryMock.findById.mockResolvedValue(null);

      await expect(service.refreshToken(refreshToken)).rejects.toThrow(
        NotAuthorizedError,
      );
    });

    it('should throw NotAuthorizedError if refresh token does not match', async () => {
      const refreshToken = 'validRefreshToken';
      const payload = { userId: '123' };
      const user: User = {
        id: '123',
        refreshToken: 'differentRefreshToken',
        email: 'user@email.com',
        name: 'User Name',
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      jwtServiceMock.verify.mockReturnValue(payload);
      userRepositoryMock.findById.mockResolvedValue(user);

      await expect(service.refreshToken(refreshToken)).rejects.toThrow(
        NotAuthorizedError,
      );
    });

    it('should throw NotAuthorizedError if refresh token is invalid', async () => {
      const refreshToken = 'invalidRefreshToken';

      jwtServiceMock.verify.mockImplementation(() => {
        throw new NotAuthorizedError({
          action: 'authenticationService.verifyRefreshToken',
          message: 'Token expirado ou inválido',
        });
      });

      await expect(service.refreshToken(refreshToken)).rejects.toThrow(
        NotAuthorizedError,
      );
    });
  });
});
