import { Test, TestingModule } from '@nestjs/testing';
import { UserService } from './user.service';
import { UserRepository } from 'src/repositories/userRepository';
import { User, UserWithPassword } from 'src/@entities/user';
import { BadRequestError } from 'src/common/applicationError';
import { mock } from 'jest-mock-extended';
import { OmitDefaultData } from 'src/utils/types';

const userRepositoryMock = mock<UserRepository>();

describe('UserService', () => {
  let service: UserService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        UserService,
        { provide: UserRepository, useValue: userRepositoryMock },
      ],
    }).compile();

    service = module.get<UserService>(UserService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should create a new user', async () => {
    const userInput: Omit<OmitDefaultData<UserWithPassword>, 'refreshToken'> = {
      name: 'John Doe',
      email: 'johndoe@example.com',
      password: 'password123',
    };

    const userCreated: User = {
      ...userInput,
      refreshToken: null,
      id: '1',
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    userRepositoryMock.create.mockResolvedValueOnce(userCreated);

    const result = await service.create(userInput);

    expect(result).toEqual(userCreated);
    expect(userRepositoryMock.create).toHaveBeenCalledTimes(1);
    expect(userRepositoryMock.create).toHaveBeenCalledWith({
      ...userInput,
      password: expect.any(String),
    });
  });

  it('should throw an error if the email already exists', async () => {
    const userInput: Omit<OmitDefaultData<UserWithPassword>, 'refreshToken'> = {
      name: 'John Doe',
      email: 'johndoe@example.com',
      password: 'password123',
    };

    userRepositoryMock.findByEmail.mockResolvedValueOnce({
      ...userInput,
      id: '1',
      createdAt: new Date(),
      updatedAt: new Date(),
      refreshToken: 'refreshToken',
    });

    await expect(service.create(userInput)).rejects.toThrow(
      new BadRequestError({
        message: 'Email já está cadastrado',
        action: 'UserService.create',
        saveLog: false,
      }),
    );

    expect(userRepositoryMock.create).not.toHaveBeenCalled();
  });
});
