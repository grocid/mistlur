package clfft

import (
	"mistlur/cl"
)

type CLFourier struct {
	Device *cl.Device
	
	logInputSize uint
	numFreqs int
	
	context *cl.Context
	commandQueue *cl.CommandQueue
	program *cl.Program
	computeCellKernel *cl.Kernel
	partialSumKernel *cl.Kernel
	averageKernel *cl.Kernel
	inputBuffer *cl.MemObject
	freqsBuffer *cl.MemObject
	realsBuffer *cl.MemObject
	imagsBuffer *cl.MemObject
	outputBuffer *cl.MemObject
}

var CoreSource = `
#define _USE_MATH_DEFINES

__kernel void compute_cell(
	__global float *input,
	__global float *freqs,
	__global float *reals,
	__global float *imags,
	const unsigned int log_input_size,
	const unsigned int count
) {
	unsigned int i = get_global_id(0);
	if (i < count) {
		unsigned int freq_idx = i >> log_input_size;
		unsigned int sample_idx = i & ((1 << log_input_size) - 1);
		float sample = input[sample_idx];
		float phase = -2 * M_PI * freqs[freq_idx] * ((float) sample_idx);
		float cos_phase, sin_phase = sincos(phase, &cos_phase);
		reals[i] = sample * cos_phase;
		imags[i] = sample * sin_phase;
	}
}
__kernel void partial_sum(
	__global float *reals,
	__global float *imags,
	const unsigned int round,
	const unsigned int count
) {
	unsigned int i = get_global_id(0);
	if (i < count) {
		unsigned int idx1 = i << round;
		unsigned int idx2 = idx1 + (1 << (round - 1));
		reals[idx1] += reals[idx2];
		imags[idx1] += imags[idx2];
	}
}
__kernel void average(
	__global float *reals,
	__global float *imags,
	__global float *output,
	const unsigned int log_input_size,
	const unsigned int count
) {
	unsigned int i = get_global_id(0);
	if (i < count) {
		unsigned int idx = i << log_input_size;
		float re = reals[idx];
		float im = imags[idx];
		output[i] = sqrt(re*re + im*im) / ((float) (1 << log_input_size));
	}
}
`

func GetDefaultDevice() (device *cl.Device, err error) {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		return nil, err
	}
	
	if len(platforms) == 0 {
		return nil, NoPlatformsError{}
	}
	
	platform := platforms[0]
	
	devices, err := platform.GetDevices(cl.DeviceTypeGPU)
	if err != nil {
		return nil, err
	}
	
	if len(devices) == 0 {
		return nil, NoDevicesError{platform.Name()}
	}
	
	device = devices[0]
	
	return device, nil
}

func New(logInputSize uint, numFreqs int) (c *CLFourier) {
	defaultDevice, _ := GetDefaultDevice()
	
	return &CLFourier{
		Device: defaultDevice,
		logInputSize: logInputSize,
		numFreqs: numFreqs,
	}
}

func (c *CLFourier) Init() (err error) {
	if c.context != nil {
		return AlreadyInitialisedError{}
	}
	
	if c.Device == nil {
		c.Device, err = GetDefaultDevice()
		if err != nil {
			return err
		}
	}
	
	// Create context
	c.context, err = cl.CreateContext([]*cl.Device{c.Device})
	if err != nil {
		c.Release()
		return err
	}
	
	// Create command queue
	c.commandQueue, err = c.context.CreateCommandQueue(c.Device, 0)
	if err != nil {
		c.Release()
		return err
	}
	
	// Create program
	c.program, err = c.context.CreateProgramWithSource([]string{CoreSource})
	if err != nil {
		c.Release()
		return err
	}
	
	// Build program
	err = c.program.BuildProgram(nil, "")
	if err != nil {
		c.Release()
		return err
	}
	
	// Create compute_cell kernel
	c.computeCellKernel, err = c.program.CreateKernel("compute_cell")
	if err != nil {
		c.Release()
		return err
	}
	
	// Create partial_sum kernel
	c.partialSumKernel, err = c.program.CreateKernel("partial_sum")
	if err != nil {
		c.Release()
		return err
	}
	
	// Create average kernel
	c.averageKernel, err = c.program.CreateKernel("average")
	if err != nil {
		c.Release()
		return err
	}
	
	inputSize := 1 << c.logInputSize
	
	// Create input buffer
	c.inputBuffer, err = c.context.CreateEmptyBuffer(cl.MemReadOnly, 4 * inputSize)
	if err != nil {
		c.Release()
		return err
	}
	
	// Create freqs buffer
	c.freqsBuffer, err = c.context.CreateEmptyBuffer(cl.MemReadOnly, 4 * c.numFreqs)
	if err != nil {
		c.Release()
		return err
	}
	
	// Create reals buffer
	c.realsBuffer, err = c.context.CreateEmptyBuffer(cl.MemReadWrite, 4 * c.numFreqs * inputSize)
	if err != nil {
		c.Release()
		return err
	}
	
	// Create imags buffer
	c.imagsBuffer, err = c.context.CreateEmptyBuffer(cl.MemReadWrite, 4 * c.numFreqs * inputSize)
	if err != nil {
		c.Release()
		return err
	}
	
	// Create output buffer
	c.outputBuffer, err = c.context.CreateEmptyBuffer(cl.MemWriteOnly, 4 * c.numFreqs)
	if err != nil {
		c.Release()
		return err
	}
	
	return nil
}

func (c *CLFourier) Release() {
	if c.outputBuffer != nil {
		c.outputBuffer.Release()
		c.outputBuffer = nil
	}
	
	if c.imagsBuffer != nil {
		c.imagsBuffer.Release()
		c.imagsBuffer = nil
	}
	
	if c.realsBuffer != nil {
		c.realsBuffer.Release()
		c.realsBuffer = nil
	}
	
	if c.freqsBuffer != nil {
		c.freqsBuffer.Release()
		c.freqsBuffer = nil
	}
	
	if c.inputBuffer != nil {
		c.inputBuffer.Release()
		c.inputBuffer = nil
	}
	
	if c.averageKernel != nil {
		c.averageKernel.Release()
		c.averageKernel = nil
	}
	
	if c.partialSumKernel != nil {
		c.partialSumKernel.Release()
		c.partialSumKernel = nil
	}
	
	if c.computeCellKernel != nil {
		c.computeCellKernel.Release()
		c.computeCellKernel = nil
	}
	
	if c.program != nil {
		c.program.Release()
		c.program = nil
	}
	
	if c.commandQueue != nil {
		c.commandQueue.Release()
		c.commandQueue = nil
	}
	
	if c.context != nil {
		c.context.Release()
		c.context = nil
	}
}

func (c *CLFourier) Sync() (err error) {
	if c.context == nil {
		return NotInitialisedError{}
	}
	
	return c.commandQueue.Finish()
}

func (c *CLFourier) WriteInput(input []float32) (err error) {
	if c.context == nil {
		return NotInitialisedError{}
	}
	
	err = c.Sync()
	if err != nil {
		return err
	}
	
	inputSize := 1 << c.logInputSize
	if len(input) != inputSize {
		return InvalidInputSizeError{inputSize, len(input)}
	}
	
	event, err := c.commandQueue.EnqueueWriteBufferFloat32(c.inputBuffer, true, 0, input, nil)
	if err != nil {
		return err
	}
	
	event.Release()
	return nil
}

func (c *CLFourier) WriteFreqs(freqs []float32) (err error) {
	if c.context == nil {
		return NotInitialisedError{}
	}
	
	err = c.Sync()
	if err != nil {
		return err
	}
	
	if len(freqs) != c.numFreqs {
		return InvalidInputSizeError{c.numFreqs, len(freqs)}
	}
	
	event, err := c.commandQueue.EnqueueWriteBufferFloat32(c.freqsBuffer, true, 0, freqs, nil)
	if err != nil {
		return err
	}
	event.Release()
	
	return nil
}

func (c *CLFourier) ReadOutput() (output []float32, err error) {
	if c.context == nil {
		return nil, NotInitialisedError{}
	}
	
	err = c.Sync()
	if err != nil {
		return nil, err
	}
	
	output = make([]float32, c.numFreqs)
	event, err := c.commandQueue.EnqueueReadBufferFloat32(c.outputBuffer, true, 0, output, nil)
	if err != nil {
		return nil, err
	}
	event.Release()
	
	return output, nil
}

func (c *CLFourier) Transform(input []float32) (output []float32, err error) {
	err = c.WriteInput(input)
	if err != nil {
		return nil, err
	}
	
	err = c.Run()
	if err != nil {
		return nil, err
	}
	
	return c.ReadOutput()
}

func (c *CLFourier) Run() (err error) {
	if c.context == nil {
		return NotInitialisedError{}
	}
	
	err = c.runComputeCell()
	if err != nil {
		return err
	}
	
	for i := uint(0); i < c.logInputSize; i++ {
		err = c.runPartialSum(i + 1)
		if err != nil {
			return err
		}
	}
	
	return c.runAverage()
}

func (c *CLFourier) runComputeCell() (err error) {
	err = c.Sync()
	if err != nil {
		return err
	}
	
	inputSize := 1 << c.logInputSize
	count := c.numFreqs * inputSize
	
	err = c.computeCellKernel.SetArgs(
		c.inputBuffer,
		c.freqsBuffer,
		c.realsBuffer,
		c.imagsBuffer,
		uint32(c.logInputSize),
		uint32(count),
	)
	
	if err != nil {
		return err
	}
	
	global, local, err := c.getWorkGroupSizes(c.computeCellKernel, count)
	if err != nil {
		return err
	}
	
	event, err := c.commandQueue.EnqueueNDRangeKernel(c.computeCellKernel, nil, []int{global}, []int{local}, nil)
	if err != nil {
		return err
	}
	event.Release()
	
	return nil
}

func (c *CLFourier) runPartialSum(round uint) (err error) {
	err = c.Sync()
	if err != nil {
		return err
	}
	
	count := c.numFreqs << (c.logInputSize - round)
	
	err = c.partialSumKernel.SetArgs(
		c.realsBuffer,
		c.imagsBuffer,
		uint32(round),
		uint32(count),
	)
	
	if err != nil {
		return err
	}
	
	global, local, err := c.getWorkGroupSizes(c.partialSumKernel, count)
	if err != nil {
		return err
	}
	
	event, err := c.commandQueue.EnqueueNDRangeKernel(c.partialSumKernel, nil, []int{global}, []int{local}, nil)
	if err != nil {
		return err
	}
	event.Release()
	
	return nil
}

func (c *CLFourier) runAverage() (err error) {
	err = c.Sync()
	if err != nil {
		return err
	}
	
	count := c.numFreqs
	
	err = c.averageKernel.SetArgs(
		c.realsBuffer,
		c.imagsBuffer,
		c.outputBuffer,
		uint32(c.logInputSize),
		uint32(count),
	)
	
	if err != nil {
		return err
	}
	
	global, local, err := c.getWorkGroupSizes(c.averageKernel, count)
	if err != nil {
		return err
	}
	
	event, err := c.commandQueue.EnqueueNDRangeKernel(c.averageKernel, nil, []int{global}, []int{local}, nil)
	if err != nil {
		return err
	}
	event.Release()
	
	return nil
}

func (c *CLFourier) getWorkGroupSizes(kernel *cl.Kernel, count int) (global, local int, err error) {
	local, err = kernel.WorkGroupSize(c.Device)
	if err != nil {
		return 0, 0, err
	}
	
	global = count
	
	// Make global a multiple of local
	d := global % local
	if d != 0 {
		global += local - d
	}
	
	return global, local, nil
}